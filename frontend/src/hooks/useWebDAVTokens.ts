import { useState, useEffect, useCallback } from 'react';
import { WebDAVAPI, WebDAVToken } from '@/lib/api/webdav';

export interface UseWebDAVTokensOptions {
  bucket?: string;
  includeExpired?: boolean;
  autoRefresh?: boolean;
  refreshInterval?: number; // in milliseconds
}

export interface UseWebDAVTokensReturn {
  tokens: WebDAVToken[];
  selectedToken: WebDAVToken | null;
  loading: boolean;
  error: string | null;
  // Actions
  loadTokens: () => Promise<void>;
  selectToken: (tokenId: number | null) => void;
  createToken: (bucket: string, expiresIn?: number) => Promise<WebDAVToken>;
  deleteToken: (tokenId: number) => Promise<void>;
  // Helpers
  isTokenExpired: (token: WebDAVToken) => boolean;
  getValidTokensForBucket: (bucket: string) => WebDAVToken[];
  getWebDAVUrl: (bucket: string) => string;
}

/**
 * Custom hook for managing WebDAV tokens
 * Provides token CRUD operations and state management
 */
export function useWebDAVTokens(options: UseWebDAVTokensOptions = {}): UseWebDAVTokensReturn {
  const {
    bucket,
    includeExpired = false,
    autoRefresh = false,
    refreshInterval = 60000 // 1 minute default
  } = options;

  const [tokens, setTokens] = useState<WebDAVToken[]>([]);
  const [selectedToken, setSelectedToken] = useState<WebDAVToken | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load tokens from API
  const loadTokens = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const tokenList = await WebDAVAPI.listTokens(bucket, includeExpired);
      setTokens(tokenList);

      // Auto-select first valid token if none selected and bucket matches
      if (bucket && !selectedToken) {
        const validToken = tokenList.find(token => 
          token.bucket === bucket && !isTokenExpired(token)
        );
        if (validToken) {
          setSelectedToken(validToken);
        }
      }
    } catch (err) {
      console.error('Load tokens error:', err);
      setError(err instanceof Error ? err.message : 'Failed to load tokens');
    } finally {
      setLoading(false);
    }
  }, [bucket, includeExpired, selectedToken]);

  // Select token by ID
  const selectToken = useCallback((tokenId: number | null) => {
    const token = tokenId ? tokens.find(t => t.id === tokenId) || null : null;
    setSelectedToken(token);
  }, [tokens]);

  // Create new token
  const createToken = useCallback(async (bucket: string, expiresIn: number = 24): Promise<WebDAVToken> => {
    try {
      setLoading(true);
      setError(null);
      const newToken = await WebDAVAPI.createToken({ bucket, expires_in: expiresIn });
      
      // Refresh token list to include new token
      await loadTokens();
      
      // Auto-select the new token
      setSelectedToken(newToken);
      
      return newToken;
    } catch (err) {
      console.error('Create token error:', err);
      const errorMsg = err instanceof Error ? err.message : 'Failed to create token';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, [loadTokens]);

  // Delete token
  const deleteToken = useCallback(async (tokenId: number) => {
    try {
      setLoading(true);
      setError(null);
      await WebDAVAPI.deleteToken(tokenId);
      
      // Remove from local state
      setTokens(prev => prev.filter(t => t.id !== tokenId));
      
      // Clear selection if deleted token was selected
      if (selectedToken && selectedToken.id === tokenId) {
        setSelectedToken(null);
      }
    } catch (err) {
      console.error('Delete token error:', err);
      const errorMsg = err instanceof Error ? err.message : 'Failed to delete token';
      setError(errorMsg);
      throw new Error(errorMsg);
    } finally {
      setLoading(false);
    }
  }, [selectedToken]);

  // Check if token is expired
  const isTokenExpired = useCallback((token: WebDAVToken): boolean => {
    return new Date(token.expires_at) < new Date();
  }, []);

  // Get valid tokens for a specific bucket
  const getValidTokensForBucket = useCallback((bucket: string): WebDAVToken[] => {
    return tokens.filter(token => 
      token.bucket === bucket && !isTokenExpired(token)
    );
  }, [tokens, isTokenExpired]);

  // Generate WebDAV URL
  const getWebDAVUrl = useCallback((bucket: string): string => {
    if (typeof window !== 'undefined') {
      return `${window.location.origin}/webdav/${bucket}`;
    }
    return `/webdav/${bucket}`;
  }, []);

  // Load tokens on mount and when bucket changes
  useEffect(() => {
    loadTokens();
  }, [loadTokens]);

  // Auto-refresh tokens if enabled
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      loadTokens();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, loadTokens]);

  // Clear selected token if it becomes invalid
  useEffect(() => {
    if (selectedToken && bucket && selectedToken.bucket !== bucket) {
      setSelectedToken(null);
    }
  }, [selectedToken, bucket]);

  return {
    tokens,
    selectedToken,
    loading,
    error,
    // Actions
    loadTokens,
    selectToken,
    createToken,
    deleteToken,
    // Helpers
    isTokenExpired,
    getValidTokensForBucket,
    getWebDAVUrl,
  };
}
