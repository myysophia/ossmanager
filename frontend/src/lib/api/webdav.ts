import apiClient from './axios';

export interface WebDAVConnectionInfo {
  url: string;
  bucket: string;
  username: string;
  supportedMethods: string[];
}

export interface WebDAVToken {
  id: number;
  bucket: string;
  token?: string;  // Only returned when creating
  expires_at: string;
  created_at: string;
}

export interface CreateTokenRequest {
  bucket: string;
  expires_in?: number;  // Hours, default 24
}

/**
 * WebDAV相关API服务
 */
export const WebDAVAPI = {
  /**
   * 创建WebDAV访问令牌
   * @param request 创建令牌请求
   * @returns 访问令牌信息
   */
  createToken: async (request: CreateTokenRequest): Promise<WebDAVToken> => {
    const response = await apiClient.post('/webdav/token', request);
    return response.data.data;
  },

  /**
   * 获取当前用户的WebDAV令牌列表
   * @param bucket 可选，按存储桶过滤
   * @param includeExpired 是否包含过期的令牌
   * @returns 令牌列表
   */
  listTokens: async (bucket?: string, includeExpired: boolean = false): Promise<WebDAVToken[]> => {
    const params = new URLSearchParams();
    if (bucket) params.append('bucket', bucket);
    if (includeExpired) params.append('include_expired', 'true');
    
    const response = await apiClient.get(`/webdav/token?${params.toString()}`);
    return response.data.data;
  },

  /**
   * 删除WebDAV访问令牌
   * @param tokenId 令牌ID
   */
  deleteToken: async (tokenId: number): Promise<void> => {
    await apiClient.delete(`/webdav/token/${tokenId}`);
  },

  /**
   * 获取WebDAV连接信息
   * @param bucket 存储桶名称
   * @returns WebDAV连接信息
   */
  getConnectionInfo: async (bucket: string): Promise<WebDAVConnectionInfo> => {
    const response = await apiClient.get(`/webdav/token/connection-info/${bucket}`);
    return response.data.data;
  },

  /**
   * 测试WebDAV连接
   * @param bucket 存储桶名称
   * @returns 连接测试结果
   */
  testConnection: async (bucket: string): Promise<any> => {
    const response = await apiClient.get(`/webdav/token/test/${bucket}`);
    return response.data.data;
  },

  /**
   * 获取用户可访问的存储桶列表
   * @returns 存储桶列表
   */
  getAccessibleBuckets: async (): Promise<string[]> => {
    const response = await apiClient.get('/oss/region-buckets/user-accessible');
    return response.data.data;
  },
};
