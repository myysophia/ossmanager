'use client';

import { OSSFile } from './api/types';
import { Toast, ToastUtils } from './toast';

/**
 * Optimistic update action types
 */
export type OptimisticAction = 'upload' | 'delete' | 'download' | 'rename' | 'update';

/**
 * Optimistic update state
 */
export interface OptimisticState {
  action: OptimisticAction;
  target: string | number; // file id or filename
  originalData?: any;
  isLoading: boolean;
  error?: string;
}

/**
 * Optimistic update manager class
 */
export class OptimisticUpdateManager {
  private pendingActions = new Map<string, OptimisticState>();
  private rollbackCallbacks = new Map<string, () => void>();

  /**
   * Generate unique key for action
   */
  private generateKey(action: OptimisticAction, target: string | number): string {
    return `${action}-${target}-${Date.now()}`;
  }

  /**
   * Start optimistic update
   */
  startUpdate(
    action: OptimisticAction,
    target: string | number,
    optimisticUpdate: () => void,
    rollbackCallback: () => void
  ): string {
    const key = this.generateKey(action, target);
    
    // Store rollback callback
    this.rollbackCallbacks.set(key, rollbackCallback);
    
    // Store optimistic state
    this.pendingActions.set(key, {
      action,
      target,
      isLoading: true,
    });

    // Apply optimistic update immediately
    optimisticUpdate();

    return key;
  }

  /**
   * Complete optimistic update (success)
   */
  completeUpdate(key: string): void {
    // Clean up
    this.pendingActions.delete(key);
    this.rollbackCallbacks.delete(key);
  }

  /**
   * Rollback optimistic update (failure)
   */
  rollbackUpdate(key: string, error?: string): void {
    const rollbackCallback = this.rollbackCallbacks.get(key);
    const state = this.pendingActions.get(key);

    if (rollbackCallback) {
      rollbackCallback();
    }

    if (state && error) {
      // Show error notification
      this.showErrorNotification(state.action, state.target, error);
    }

    // Clean up
    this.pendingActions.delete(key);
    this.rollbackCallbacks.delete(key);
  }

  /**
   * Check if action is pending
   */
  isPending(action: OptimisticAction, target: string | number): boolean {
    return Array.from(this.pendingActions.values()).some(
      state => state.action === action && state.target === target && state.isLoading
    );
  }

  /**
   * Get pending actions
   */
  getPendingActions(): OptimisticState[] {
    return Array.from(this.pendingActions.values());
  }

  /**
   * Show error notification based on action type
   */
  private showErrorNotification(action: OptimisticAction, target: string | number, error: string): void {
    switch (action) {
      case 'upload':
        ToastUtils.uploadError(target.toString(), error);
        break;
      case 'delete':
        ToastUtils.deleteError(target.toString(), error);
        break;
      case 'download':
        ToastUtils.downloadError(target.toString(), error);
        break;
      case 'rename':
        Toast.error({
          title: '重命名失败',
          description: error || `无法重命名文件`,
        });
        break;
      case 'update':
        Toast.error({
          title: '更新失败',
          description: error || '更新操作失败',
        });
        break;
      default:
        Toast.error({
          title: '操作失败',
          description: error || '未知错误',
        });
    }
  }
}

/**
 * Global optimistic update manager instance
 */
export const optimisticUpdateManager = new OptimisticUpdateManager();

/**
 * Optimistic update utilities for common file operations
 */
export const OptimisticUtils = {
  /**
   * Optimistic file upload
   */
  uploadFile: (
    file: File,
    currentFiles: OSSFile[],
    setFiles: (files: OSSFile[] | ((prev: OSSFile[]) => OSSFile[])) => void
  ): string => {
    // Create optimistic file entry
    const optimisticFile: OSSFile = {
      id: -Date.now(), // Temporary negative ID
      file_name: file.name,
      original_filename: file.name,
      file_size: file.size,
      file_type: file.type,
      storage_type: 'ALIYUN_OSS',
      object_key: `temp/${file.name}`,
      config_id: 0,
      config_name: 'uploading',
      md5: '',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    const key = optimisticUpdateManager.startUpdate(
      'upload',
      file.name,
      () => {
        // Optimistically add the file to the list
        setFiles(prev => [optimisticFile, ...prev]);
        ToastUtils.uploadProgress(file.name, 0);
      },
      () => {
        // Rollback: remove the optimistic file
        setFiles(prev => prev.filter(f => f.id !== optimisticFile.id));
      }
    );

    return key;
  },

  /**
   * Optimistic file deletion
   */
  deleteFile: (
    fileId: number,
    currentFiles: OSSFile[],
    setFiles: (files: OSSFile[] | ((prev: OSSFile[]) => OSSFile[])) => void
  ): string => {
    const fileToDelete = currentFiles.find(f => f.id === fileId);
    const filename = fileToDelete ? decodeURIComponent(fileToDelete.original_filename) : '';

    const key = optimisticUpdateManager.startUpdate(
      'delete',
      fileId,
      () => {
        // Optimistically remove the file from the list
        setFiles(prev => prev.filter(f => f.id !== fileId));
        if (filename) {
          ToastUtils.deleteSuccess(filename);
        }
      },
      () => {
        // Rollback: restore the file to the list
        if (fileToDelete) {
          setFiles(prev => {
            const exists = prev.find(f => f.id === fileId);
            if (!exists) {
              return [...prev, fileToDelete].sort((a, b) => 
                new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
              );
            }
            return prev;
          });
        }
      }
    );

    return key;
  },

  /**
   * Optimistic batch file deletion
   */
  deleteFiles: (
    fileIds: number[],
    currentFiles: OSSFile[],
    setFiles: (files: OSSFile[] | ((prev: OSSFile[]) => OSSFile[])) => void
  ): string => {
    const filesToDelete = currentFiles.filter(f => fileIds.includes(f.id));

    const key = optimisticUpdateManager.startUpdate(
      'delete',
      fileIds.join(','),
      () => {
        // Optimistically remove files from the list
        setFiles(prev => prev.filter(f => !fileIds.includes(f.id)));
        Toast.success({
          title: '批量删除成功',
          description: `已删除 ${fileIds.length} 个文件`,
        });
      },
      () => {
        // Rollback: restore all files to the list
        setFiles(prev => {
          const restored = [...prev];
          filesToDelete.forEach(file => {
            const exists = restored.find(f => f.id === file.id);
            if (!exists) {
              restored.push(file);
            }
          });
          return restored.sort((a, b) => 
            new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
          );
        });
      }
    );

    return key;
  },

  /**
   * Complete upload operation
   */
  completeUpload: (
    key: string,
    uploadedFile: OSSFile,
    setFiles: (files: OSSFile[] | ((prev: OSSFile[]) => OSSFile[])) => void,
    tempId?: number
  ): void => {
    // Replace the optimistic file with the actual uploaded file
    setFiles(prev => prev.map(f => 
      (tempId && f.id === tempId) ? uploadedFile : f
    ));

    // Complete the optimistic update
    optimisticUpdateManager.completeUpdate(key);
    
    ToastUtils.uploadSuccess(uploadedFile.original_filename);
  },

  /**
   * Complete delete operation
   */
  completeDelete: (key: string): void => {
    optimisticUpdateManager.completeUpdate(key);
  },

  /**
   * Fail operation
   */
  failOperation: (key: string, error: string): void => {
    optimisticUpdateManager.rollbackUpdate(key, error);
  },

  /**
   * Handle API error and rollback
   */
  handleError: (key: string, error: any): void => {
    const errorMessage = error?.response?.data?.message || error?.message || '操作失败';
    optimisticUpdateManager.rollbackUpdate(key, errorMessage);
  },
};
