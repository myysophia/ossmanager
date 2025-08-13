import axios, { AxiosProgressEvent } from 'axios';
import apiClient from './axios';

/**
 * WebDAV Browser API Usage Examples
 * 
 * 1. List objects in a directory:
 * ```typescript
 * const files = await WebDAVBrowserAPI.listObjects('my-bucket', '/documents');
 * console.log(`Found ${files.total} items:`, files.items);
 * ```
 * 
 * 2. Upload a file with progress tracking:
 * ```typescript
 * const file = document.querySelector('input[type="file"]').files[0];
 * await WebDAVBrowserAPI.uploadFile(
 *   'my-bucket',
 *   '/documents/upload.txt',
 *   file,
 *   (progress) => {
 *     const percent = Math.round((progress.loaded * 100) / progress.total);
 *     console.log(`Upload progress: ${percent}%`);
 *   }
 * );
 * ```
 * 
 * 3. Create a new folder:
 * ```typescript
 * await WebDAVBrowserAPI.createFolder('my-bucket', '/documents/new-folder');
 * ```
 * 
 * 4. Rename/move a file:
 * ```typescript
 * await WebDAVBrowserAPI.renameObject(
 *   'my-bucket',
 *   '/documents/old-name.txt',
 *   '/documents/new-name.txt'
 * );
 * ```
 * 
 * 5. Delete a file or folder:
 * ```typescript
 * await WebDAVBrowserAPI.deleteObject('my-bucket', '/documents/file-to-delete.txt');
 * ```
 * 
 * 6. Using the class-based client directly:
 * ```typescript
 * import { webdavBrowserClient } from '@/lib/api/webdavBrowser';
 * 
 * const files = await webdavBrowserClient.listObjects({
 *   bucket: 'my-bucket',
 *   path: '/documents'
 * });
 * ```
 */

/**
 * WebDAV 文件对象接口
 */
export interface WebDAVFileObject {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  mtime: string;
  mimeType?: string;
}

/**
 * WebDAV 列表响应接口
 */
export interface WebDAVListResponse {
  items: WebDAVFileObject[];
  path: string;
  total: number;
}

/**
 * WebDAV 操作响应接口
 */
export interface WebDAVOperationResponse {
  success: boolean;
  message: string;
  path: string;
}

/**
 * 文件上传请求参数
 */
export interface UploadFileRequest {
  bucket: string;
  path: string;
  file: File;
  onProgress?: (progress: AxiosProgressEvent) => void;
}

/**
 * 重命名对象请求参数
 */
export interface RenameObjectRequest {
  bucket: string;
  oldPath: string;
  newPath: string;
}

/**
 * 创建文件夹请求参数
 */
export interface CreateFolderRequest {
  bucket: string;
  path: string;
}

/**
 * 删除对象请求参数
 */
export interface DeleteObjectRequest {
  bucket: string;
  path: string;
}

/**
 * 列出对象请求参数
 */
export interface ListObjectsRequest {
  bucket: string;
  path?: string;
}

/**
 * WebDAV 浏览器客户端服务
 * 提供与 WebDAV 代理端点交互的统一接口
 */
export class WebDAVBrowserClient {
  private baseURL: string;

  constructor() {
    this.baseURL = '/webdav/objects';
  }

  /**
   * 获取认证头部
   * 自动从 localStorage 获取 JWT token
   */
  private getAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {};
    
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
    }
    
    return headers;
  }

  /**
   * 列出目录中的对象
   * 映射到 WebDAV PROPFIND 操作
   * 
   * @param request 列出对象请求参数
   * @returns WebDAV 列表响应
   */
  async listObjects(request: ListObjectsRequest): Promise<WebDAVListResponse> {
    try {
      const params = new URLSearchParams();
      if (request.path && request.path !== '/') {
        params.append('path', request.path);
      }

      const url = `${this.baseURL}/${request.bucket}${params.toString() ? `?${params.toString()}` : ''}`;
      
      const response = await apiClient.get(url);
      
      // 响应已经通过 axios 拦截器处理，直接返回 data 字段
      return response.data || response;
    } catch (error) {
      console.error('列出对象失败:', error);
      throw error;
    }
  }

  /**
   * 上传文件
   * 映射到 WebDAV PUT 操作
   * 支持上传进度监听
   * 
   * @param request 上传文件请求参数
   * @returns WebDAV 操作响应
   */
  async uploadFile(request: UploadFileRequest): Promise<WebDAVOperationResponse> {
    try {
      const params = new URLSearchParams();
      params.append('path', request.path);

      const url = `${this.baseURL}/${request.bucket}/file?${params.toString()}`;
      
      // 创建独立的 axios 实例以支持上传进度
      const uploadClient = axios.create({
        baseURL: apiClient.defaults.baseURL,
        timeout: apiClient.defaults.timeout,
        headers: {
          ...this.getAuthHeaders(),
          'Content-Type': 'application/octet-stream',
        },
      });

      const response = await uploadClient.post(url, request.file, {
        onUploadProgress: request.onProgress,
        headers: {
          ...this.getAuthHeaders(),
          'Content-Type': 'application/octet-stream',
        },
      });

      // 处理响应数据
      let responseData = response.data;
      
      // 检查是否是标准的API响应格式
      if (responseData && typeof responseData === 'object') {
        if (responseData.code !== undefined) {
          if (responseData.code !== 0 && responseData.code !== 200) {
            throw new Error(responseData.message || '上传失败');
          }
          responseData = responseData.data;
        }
      }

      return responseData || {
        success: true,
        message: '文件上传成功',
        path: request.path,
      };
    } catch (error) {
      console.error('文件上传失败:', error);
      throw error;
    }
  }

  /**
   * 删除对象（文件或文件夹）
   * 映射到 WebDAV DELETE 操作
   * 
   * @param request 删除对象请求参数
   * @returns WebDAV 操作响应
   */
  async deleteObject(request: DeleteObjectRequest): Promise<WebDAVOperationResponse> {
    try {
      const params = new URLSearchParams();
      params.append('path', request.path);

      const url = `${this.baseURL}/${request.bucket}?${params.toString()}`;
      
      const response = await apiClient.delete(url);
      
      // 响应已经通过 axios 拦截器处理，直接返回 data 字段
      return response.data || response;
    } catch (error) {
      console.error('删除对象失败:', error);
      throw error;
    }
  }

  /**
   * 重命名/移动对象
   * 映射到 WebDAV MOVE 操作
   * 
   * @param request 重命名对象请求参数
   * @returns WebDAV 操作响应
   */
  async renameObject(request: RenameObjectRequest): Promise<WebDAVOperationResponse> {
    try {
      const url = `${this.baseURL}/${request.bucket}/rename`;
      
      const requestBody = {
        oldPath: request.oldPath,
        newPath: request.newPath,
      };
      
      const response = await apiClient.patch(url, requestBody);
      
      // 响应已经通过 axios 拦截器处理，直接返回 data 字段
      return response.data || response;
    } catch (error) {
      console.error('重命名对象失败:', error);
      throw error;
    }
  }

  /**
   * 创建文件夹
   * 映射到 WebDAV MKCOL 操作
   * 
   * @param request 创建文件夹请求参数
   * @returns WebDAV 操作响应
   */
  async createFolder(request: CreateFolderRequest): Promise<WebDAVOperationResponse> {
    try {
      const url = `${this.baseURL}/${request.bucket}/mkdir`;
      
      const requestBody = {
        path: request.path,
      };
      
      const response = await apiClient.post(url, requestBody);
      
      // 响应已经通过 axios 拦截器处理，直接返回 data 字段
      return response.data || response;
    } catch (error) {
      console.error('创建文件夹失败:', error);
      throw error;
    }
  }
}

/**
 * WebDAV 浏览器客户端单例实例
 */
export const webdavBrowserClient = new WebDAVBrowserClient();

/**
 * WebDAV 浏览器 API - 提供便捷的函数式调用接口
 */
export const WebDAVBrowserAPI = {
  /**
   * 列出目录中的对象
   * @param bucket 存储桶名称
   * @param path 目录路径（可选，默认为根目录）
   * @returns WebDAV 列表响应
   */
  listObjects: (bucket: string, path?: string): Promise<WebDAVListResponse> => {
    return webdavBrowserClient.listObjects({ bucket, path });
  },

  /**
   * 上传文件
   * @param bucket 存储桶名称
   * @param path 文件路径
   * @param file 要上传的文件
   * @param onProgress 上传进度回调函数（可选）
   * @returns WebDAV 操作响应
   */
  uploadFile: (
    bucket: string,
    path: string,
    file: File,
    onProgress?: (progress: AxiosProgressEvent) => void
  ): Promise<WebDAVOperationResponse> => {
    return webdavBrowserClient.uploadFile({ bucket, path, file, onProgress });
  },

  /**
   * 删除对象（文件或文件夹）
   * @param bucket 存储桶名称
   * @param path 对象路径
   * @returns WebDAV 操作响应
   */
  deleteObject: (bucket: string, path: string): Promise<WebDAVOperationResponse> => {
    return webdavBrowserClient.deleteObject({ bucket, path });
  },

  /**
   * 重命名/移动对象
   * @param bucket 存储桶名称
   * @param oldPath 原路径
   * @param newPath 新路径
   * @returns WebDAV 操作响应
   */
  renameObject: (bucket: string, oldPath: string, newPath: string): Promise<WebDAVOperationResponse> => {
    return webdavBrowserClient.renameObject({ bucket, oldPath, newPath });
  },

  /**
   * 创建文件夹
   * @param bucket 存储桶名称
   * @param path 文件夹路径
   * @returns WebDAV 操作响应
   */
  createFolder: (bucket: string, path: string): Promise<WebDAVOperationResponse> => {
    return webdavBrowserClient.createFolder({ bucket, path });
  },
};

// 默认导出 API 对象以保持与其他 API 文件的一致性
export default WebDAVBrowserAPI;
