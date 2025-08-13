'use client';

import { createStandaloneToast } from '@chakra-ui/react';

const { toast } = createStandaloneToast();

/**
 * Toast notification types
 */
export type ToastType = 'success' | 'error' | 'warning' | 'info' | 'loading';

export interface ToastOptions {
  title: string;
  description?: string;
  duration?: number;
  isClosable?: boolean;
}

/**
 * Centralized toast notification system
 */
export const Toast = {
  /**
   * Show success toast
   */
  success: (options: ToastOptions) => {
    return toast({
      status: 'success',
      duration: 3000,
      isClosable: true,
      position: 'top-right',
      ...options,
    });
  },

  /**
   * Show error toast
   */
  error: (options: ToastOptions) => {
    return toast({
      status: 'error',
      duration: 5000,
      isClosable: true,
      position: 'top-right',
      ...options,
    });
  },

  /**
   * Show warning toast
   */
  warning: (options: ToastOptions) => {
    return toast({
      status: 'warning',
      duration: 4000,
      isClosable: true,
      position: 'top-right',
      ...options,
    });
  },

  /**
   * Show info toast
   */
  info: (options: ToastOptions) => {
    return toast({
      status: 'info',
      duration: 3000,
      isClosable: true,
      position: 'top-right',
      ...options,
    });
  },

  /**
   * Show loading toast
   */
  loading: (options: ToastOptions) => {
    return toast({
      status: 'loading',
      duration: null,
      isClosable: false,
      position: 'top-right',
      ...options,
    });
  },

  /**
   * Close specific toast
   */
  close: (id: string | number) => {
    toast.close(id);
  },

  /**
   * Close all toasts
   */
  closeAll: () => {
    toast.closeAll();
  },

  /**
   * Update existing toast
   */
  update: (id: string | number, options: Partial<ToastOptions> & { status?: ToastType }) => {
    toast.update(id, {
      duration: 3000,
      isClosable: true,
      position: 'top-right',
      ...options,
    });
  },
};

/**
 * Toast utilities for common scenarios
 */
export const ToastUtils = {
  /**
   * Show upload success toast
   */
  uploadSuccess: (filename: string) => {
    Toast.success({
      title: '上传成功',
      description: `文件 "${filename}" 已成功上传`,
    });
  },

  /**
   * Show upload error toast
   */
  uploadError: (filename: string, error?: string) => {
    Toast.error({
      title: '上传失败',
      description: error || `文件 "${filename}" 上传失败`,
    });
  },

  /**
   * Show delete success toast
   */
  deleteSuccess: (filename?: string) => {
    Toast.success({
      title: '删除成功',
      description: filename ? `文件 "${filename}" 已删除` : '文件已删除',
    });
  },

  /**
   * Show delete error toast
   */
  deleteError: (filename?: string, error?: string) => {
    Toast.error({
      title: '删除失败',
      description: error || (filename ? `无法删除文件 "${filename}"` : '删除操作失败'),
    });
  },

  /**
   * Show download success toast
   */
  downloadSuccess: (filename: string) => {
    Toast.success({
      title: '下载开始',
      description: `文件 "${filename}" 开始下载`,
    });
  },

  /**
   * Show download error toast
   */
  downloadError: (filename?: string, error?: string) => {
    Toast.error({
      title: '下载失败',
      description: error || (filename ? `无法下载文件 "${filename}"` : '下载失败'),
    });
  },

  /**
   * Show network error toast
   */
  networkError: (operation?: string) => {
    Toast.error({
      title: '网络错误',
      description: operation ? `${operation}失败，请检查网络连接` : '网络连接异常，请稍后重试',
    });
  },

  /**
   * Show permission error toast
   */
  permissionError: (operation?: string) => {
    Toast.error({
      title: '权限不足',
      description: operation ? `没有权限执行${operation}操作` : '您没有执行此操作的权限',
    });
  },

  /**
   * Show generic loading toast
   */
  loading: (message: string) => {
    return Toast.loading({
      title: message,
    });
  },

  /**
   * Show upload progress toast
   */
  uploadProgress: (filename: string, progress: number) => {
    const id = `upload-${filename}`;
    if (progress === 0) {
      return toast({
        id,
        status: 'loading',
        title: '正在上传',
        description: `文件 "${filename}" (0%)`,
        duration: null,
        isClosable: false,
        position: 'top-right',
      });
    } else if (progress === 100) {
      Toast.update(id, {
        status: 'success',
        title: '上传完成',
        description: `文件 "${filename}" 上传成功`,
        duration: 3000,
        isClosable: true,
      });
    } else {
      Toast.update(id, {
        description: `文件 "${filename}" (${progress}%)`,
      });
    }
    return id;
  },
};
