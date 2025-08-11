import apiClient from './axios';

export interface WebDAVConnectionInfo {
  url: string;
  bucket: string;
  username: string;
  supportedMethods: string[];
}

export interface WebDAVAccessToken {
  token: string;
  expires_at: string;
  bucket: string;
}

/**
 * WebDAV相关API服务
 */
export const WebDAVAPI = {
  /**
   * 获取WebDAV连接信息
   * @param bucket 存储桶名称
   * @returns WebDAV连接信息
   */
  getConnectionInfo: async (bucket: string): Promise<WebDAVConnectionInfo> => {
    const response = await apiClient.get(`/webdav/connection-info/${bucket}`);
    return response.data;
  },

  /**
   * 生成WebDAV访问令牌
   * @param bucket 存储桶名称
   * @returns 访问令牌信息
   */
  generateAccessToken: async (bucket: string): Promise<WebDAVAccessToken> => {
    const response = await apiClient.post(`/webdav/generate-token/${bucket}`);
    return response.data;
  },

  /**
   * 测试WebDAV连接
   * @param bucket 存储桶名称
   * @returns 连接是否成功
   */
  testConnection: async (bucket: string): Promise<boolean> => {
    try {
      await apiClient.get(`/webdav/test/${bucket}`);
      return true;
    } catch {
      return false;
    }
  },

  /**
   * 获取用户可访问的存储桶列表
   * @returns 存储桶列表
   */
  getAccessibleBuckets: async (): Promise<string[]> => {
    const response = await apiClient.get('/webdav/buckets');
    return response.data;
  },

  /**
   * 撤销WebDAV访问令牌
   * @param token 访问令牌
   */
  revokeAccessToken: async (token: string): Promise<void> => {
    await apiClient.delete(`/webdav/tokens/${token}`);
  },

  /**
   * 获取WebDAV使用统计
   * @param bucket 存储桶名称
   * @returns 使用统计信息
   */
  getUsageStats: async (bucket: string) => {
    const response = await apiClient.get(`/webdav/stats/${bucket}`);
    return response.data;
  },
};
