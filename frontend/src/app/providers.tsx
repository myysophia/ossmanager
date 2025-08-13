'use client';

import React from 'react';
import { CacheProvider } from '@chakra-ui/next-js';
import { ChakraProvider } from '@chakra-ui/react';
import { SWRConfig } from 'swr';
import theme from '../theme';
import { FileAPI } from '../lib/api';

// SWR default fetcher
const fetcher = async (url: string) => {
  // Handle different SWR key formats
  if (url.startsWith('/api/files')) {
    const params = new URLSearchParams(url.split('?')[1]);
    const queryParams: any = {};
    
    params.forEach((value, key) => {
      queryParams[key] = value;
    });
    
    return FileAPI.getFiles(queryParams);
  }
  
  throw new Error(`Unknown SWR key: ${url}`);
};

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <CacheProvider>
      <ChakraProvider theme={theme}>
        <SWRConfig 
          value={{
            fetcher,
            revalidateOnFocus: false,
            revalidateOnReconnect: true,
            errorRetryCount: 3,
            errorRetryInterval: 1000,
            dedupingInterval: 5000,
            focusThrottleInterval: 5000,
            onError: (error) => {
              console.error('SWR Error:', error);
            },
          }}
        >
          {children}
        </SWRConfig>
      </ChakraProvider>
    </CacheProvider>
  );
}
