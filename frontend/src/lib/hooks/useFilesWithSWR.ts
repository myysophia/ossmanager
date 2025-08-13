'use client';

import useSWR, { mutate } from 'swr';
import { useCallback } from 'react';
import { FileAPI, OSSFile, FileQueryParams, PageResponse } from '../api';
import { OptimisticUtils } from '../optimistic-updates';
import { ToastUtils } from '../toast';

/**
 * SWR key generator for files
 */
const getFilesKey = (params?: FileQueryParams) => {
  if (!params) return '/api/files';
  const searchParams = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      searchParams.append(key, value.toString());
    }
  });
  return `/api/files?${searchParams.toString()}`;
};

/**
 * File management hook with SWR for caching and revalidation
 */
export const useFilesWithSWR = (params?: FileQueryParams) => {
  const key = getFilesKey(params);
  
  const { data, error, isLoading, mutate: revalidate } = useSWR(
    key,
    () => FileAPI.getFiles(params),
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
      errorRetryCount: 3,
      errorRetryInterval: 1000,
      dedupingInterval: 5000, // Avoid duplicate requests within 5 seconds
    }
  );

  /**
   * Optimistic upload with SWR cache update
   */
  const uploadFile = useCallback(async (
    file: File, 
    options?: { regionCode?: string; bucketName?: string }
  ) => {
    const currentData = data as PageResponse<OSSFile>;
    if (!currentData) return;

    const currentFiles = currentData.items || [];
    
    // Start optimistic update
    const updateKey = OptimisticUtils.uploadFile(file, currentFiles, (updatedFiles) => {
      mutate(key, {
        ...currentData,
        items: Array.isArray(updatedFiles) ? updatedFiles : updatedFiles(currentFiles),
        total: currentData.total + 1,
      }, false);
    });

    try {
      // Perform actual upload
      const uploadedFile = await FileAPI.uploadFile(file, options);
      
      // Update SWR cache with real data
      const updatedData = await revalidate();
      
      // Complete optimistic update
      OptimisticUtils.completeUpload(
        updateKey,
        uploadedFile as unknown as OSSFile,
        () => {}, // No local state update needed since SWR handles it
        -Date.now()
      );

      return uploadedFile;
    } catch (error) {
      // Handle error and rollback
      OptimisticUtils.handleError(updateKey, error);
      
      // Revalidate to restore correct state
      revalidate();
      
      throw error;
    }
  }, [data, key, revalidate]);

  /**
   * Optimistic delete with SWR cache update
   */
  const deleteFile = useCallback(async (fileId: number) => {
    const currentData = data as PageResponse<OSSFile>;
    if (!currentData) return;

    const currentFiles = currentData.items || [];
    const fileToDelete = currentFiles.find(f => f.id === fileId);
    
    // Start optimistic update
    const updateKey = OptimisticUtils.deleteFile(fileId, currentFiles, (updatedFiles) => {
      const newItems = Array.isArray(updatedFiles) ? updatedFiles : updatedFiles(currentFiles);
      mutate(key, {
        ...currentData,
        items: newItems,
        total: Math.max(0, currentData.total - 1),
      }, false);
    });

    try {
      // Perform actual deletion
      await FileAPI.deleteFile(fileId);
      
      // Revalidate to get fresh data
      await revalidate();
      
      // Complete optimistic update
      OptimisticUtils.completeDelete(updateKey);
      
      return true;
    } catch (error) {
      // Handle error and rollback
      OptimisticUtils.handleError(updateKey, error);
      
      // Revalidate to restore correct state
      revalidate();
      
      throw error;
    }
  }, [data, key, revalidate]);

  /**
   * Batch delete with optimistic updates
   */
  const deleteFiles = useCallback(async (fileIds: number[]) => {
    const currentData = data as PageResponse<OSSFile>;
    if (!currentData || fileIds.length === 0) return;

    const currentFiles = currentData.items || [];
    
    // Start optimistic update
    const updateKey = OptimisticUtils.deleteFiles(fileIds, currentFiles, (updatedFiles) => {
      const newItems = Array.isArray(updatedFiles) ? updatedFiles : updatedFiles(currentFiles);
      mutate(key, {
        ...currentData,
        items: newItems,
        total: Math.max(0, currentData.total - fileIds.length),
      }, false);
    });

    try {
      // Perform actual batch deletion (if API supports it, otherwise delete one by one)
      await Promise.all(fileIds.map(id => FileAPI.deleteFile(id)));
      
      // Revalidate to get fresh data
      await revalidate();
      
      // Complete optimistic update
      OptimisticUtils.completeDelete(updateKey);
      
      return true;
    } catch (error) {
      // Handle error and rollback
      OptimisticUtils.handleError(updateKey, error);
      
      // Revalidate to restore correct state
      revalidate();
      
      throw error;
    }
  }, [data, key, revalidate]);

  /**
   * Get download URL
   */
  const getDownloadUrl = useCallback(async (fileId: number, expireHours?: number) => {
    try {
      const response = await FileAPI.getFileDownloadURL(fileId, expireHours);
      return response.download_url;
    } catch (error) {
      const file = (data as PageResponse<OSSFile>)?.items?.find(f => f.id === fileId);
      const filename = file ? decodeURIComponent(file.original_filename) : '';
      ToastUtils.downloadError(filename, error instanceof Error ? error.message : '获取下载链接失败');
      throw error;
    }
  }, [data]);

  /**
   * Download file with toast notifications
   */
  const downloadFile = useCallback(async (fileId: number) => {
    const currentData = data as PageResponse<OSSFile>;
    const file = currentData?.items?.find(f => f.id === fileId);
    const filename = file ? decodeURIComponent(file.original_filename) : `file_${fileId}`;

    try {
      const downloadUrl = await getDownloadUrl(fileId);
      
      // Create download link
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename;
      link.style.display = 'none';
      
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      ToastUtils.downloadSuccess(filename);
      
      return true;
    } catch (error) {
      ToastUtils.downloadError(filename, error instanceof Error ? error.message : '下载失败');
      throw error;
    }
  }, [data, getDownloadUrl]);

  /**
   * Refresh files data
   */
  const refreshFiles = useCallback(async () => {
    return revalidate();
  }, [revalidate]);

  /**
   * Invalidate and refetch files
   */
  const invalidateFiles = useCallback(async () => {
    // Clear all file-related cache
    mutate(
      (key) => typeof key === 'string' && key.startsWith('/api/files'),
      undefined,
      { revalidate: true }
    );
  }, []);

  return {
    // Data
    files: data?.items || [],
    total: data?.total || 0,
    loading: isLoading,
    error,
    
    // Actions
    uploadFile,
    deleteFile,
    deleteFiles,
    downloadFile,
    getDownloadUrl,
    
    // Cache management
    refreshFiles,
    invalidateFiles,
    revalidate,
  };
};

/**
 * Hook for managing single file operations
 */
export const useFileWithSWR = (fileId: number | null) => {
  const key = fileId ? `/api/files/${fileId}` : null;
  
  const { data, error, isLoading, mutate: revalidate } = useSWR(
    key,
    async () => {
      if (!fileId) return null;
      // If there's no dedicated single file API, we can derive it from the list
      const filesResponse = await FileAPI.getFiles({ page: 1, page_size: 1000 });
      return filesResponse.items.find(f => f.id === fileId) || null;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: true,
    }
  );

  return {
    file: data,
    loading: isLoading,
    error,
    revalidate,
  };
};

/**
 * Global file cache utilities
 */
export const FileCache = {
  /**
   * Update file in all relevant caches
   */
  updateFile: (updatedFile: OSSFile) => {
    mutate(
      (key) => typeof key === 'string' && key.startsWith('/api/files'),
      (data: PageResponse<OSSFile> | undefined) => {
        if (!data) return data;
        return {
          ...data,
          items: data.items.map(f => f.id === updatedFile.id ? updatedFile : f),
        };
      },
      false
    );
  },

  /**
   * Remove file from all relevant caches
   */
  removeFile: (fileId: number) => {
    mutate(
      (key) => typeof key === 'string' && key.startsWith('/api/files'),
      (data: PageResponse<OSSFile> | undefined) => {
        if (!data) return data;
        return {
          ...data,
          items: data.items.filter(f => f.id !== fileId),
          total: Math.max(0, data.total - 1),
        };
      },
      false
    );
  },

  /**
   * Add file to all relevant caches
   */
  addFile: (newFile: OSSFile) => {
    mutate(
      (key) => typeof key === 'string' && key.startsWith('/api/files'),
      (data: PageResponse<OSSFile> | undefined) => {
        if (!data) return data;
        return {
          ...data,
          items: [newFile, ...data.items],
          total: data.total + 1,
        };
      },
      false
    );
  },

  /**
   * Invalidate all file caches
   */
  invalidateAll: () => {
    mutate(
      (key) => typeof key === 'string' && key.startsWith('/api/files'),
      undefined,
      { revalidate: true }
    );
  },
};
