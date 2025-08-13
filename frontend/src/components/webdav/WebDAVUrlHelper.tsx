'use client';

import React from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Code,
  Alert,
  AlertIcon,
  useToast,
  IconButton,
  Tooltip,
  Flex,
  Spacer,
} from '@chakra-ui/react';
import { CopyIcon, ExternalLinkIcon, InfoIcon } from '@chakra-ui/icons';
import { WebDAVToken } from '@/lib/api/webdav';

export interface WebDAVUrlHelperProps {
  selectedBucket: string;
  selectedToken?: WebDAVToken | null;
  size?: 'sm' | 'md' | 'lg';
  showInstructions?: boolean;
}

export default function WebDAVUrlHelper({
  selectedBucket,
  selectedToken,
  size = 'md',
  showInstructions = true
}: WebDAVUrlHelperProps) {
  const toast = useToast();

  const copyToClipboard = async (text: string, label: string = '内容') => {
    try {
      await navigator.clipboard.writeText(text);
      toast({
        title: `${label}已复制到剪贴板`,
        status: 'success',
        duration: 2000,
        isClosable: true,
      });
    } catch (error) {
      toast({
        title: '复制失败',
        status: 'error',
        duration: 2000,
        isClosable: true,
      });
    }
  };

  const getWebDAVUrl = (bucket: string) => {
    if (typeof window !== 'undefined') {
      return `${window.location.origin}/webdav/${bucket}`;
    }
    return `/webdav/${bucket}`;
  };

  const openInExternalClient = () => {
    const url = getWebDAVUrl(selectedBucket);
    // Try to open with file:// protocol which might trigger OS file manager
    window.open(url, '_blank');
  };

  if (!selectedBucket) {
    return (
      <Alert status="info" size={size}>
        <AlertIcon />
        <Text fontSize={size === 'sm' ? 'sm' : 'md'}>
          请先选择存储桶
        </Text>
      </Alert>
    );
  }

  const webdavUrl = getWebDAVUrl(selectedBucket);
  const isTokenExpired = selectedToken ? new Date(selectedToken.expires_at) < new Date() : false;

  return (
    <Card size={size}>
      <CardHeader pb={2}>
        <Flex align="center">
          <HStack>
            <ExternalLinkIcon />
            <Heading size="sm">外部WebDAV客户端</Heading>
          </HStack>
          <Spacer />
          <Tooltip label="在外部客户端中打开">
            <IconButton
              aria-label="打开WebDAV"
              icon={<ExternalLinkIcon />}
              size="sm"
              variant="outline"
              onClick={openInExternalClient}
              isDisabled={!selectedBucket}
            />
          </Tooltip>
        </Flex>
      </CardHeader>
      <CardBody pt={0}>
        <VStack spacing={4} align="stretch">
          {/* WebDAV URL */}
          <Box>
            <Text fontSize="sm" fontWeight="medium" mb={2}>
              WebDAV 地址:
            </Text>
            <HStack>
              <Code
                fontSize="sm"
                p={2}
                borderRadius="md"
                bg="gray.50"
                flex={1}
                wordBreak="break-all"
              >
                {webdavUrl}
              </Code>
              <Tooltip label="复制地址">
                <IconButton
                  aria-label="复制WebDAV地址"
                  icon={<CopyIcon />}
                  size="sm"
                  variant="ghost"
                  onClick={() => copyToClipboard(webdavUrl, 'WebDAV地址')}
                />
              </Tooltip>
            </HStack>
          </Box>

          {/* Token Status */}
          {selectedToken ? (
            <Alert status={isTokenExpired ? 'error' : 'success'} size="sm">
              <AlertIcon />
              <VStack align="start" spacing={1}>
                <Text fontSize="sm" fontWeight="medium">
                  {isTokenExpired ? '令牌已过期' : '令牌可用'}
                </Text>
                {!isTokenExpired && (
                  <Text fontSize="xs">
                    令牌到期时间: {new Date(selectedToken.expires_at).toLocaleString()}
                  </Text>
                )}
                {isTokenExpired && (
                  <Text fontSize="xs">
                    请创建新的访问令牌以继续使用外部客户端
                  </Text>
                )}
              </VStack>
            </Alert>
          ) : (
            <Alert status="warning" size="sm">
              <AlertIcon />
              <VStack align="start" spacing={1}>
                <Text fontSize="sm" fontWeight="medium">
                  未选择访问令牌
                </Text>
                <Text fontSize="xs">
                  外部WebDAV客户端需要访问令牌进行身份验证
                </Text>
              </VStack>
            </Alert>
          )}

          {/* Quick Actions */}
          <VStack spacing={2} align="stretch">
            <Text fontSize="sm" fontWeight="medium">
              快速操作:
            </Text>
            <HStack>
              <Button
                size="sm"
                leftIcon={<CopyIcon />}
                variant="outline"
                onClick={() => copyToClipboard(webdavUrl, 'WebDAV地址')}
              >
                复制地址
              </Button>
              {selectedToken && !isTokenExpired && (
                <Button
                  size="sm"
                  leftIcon={<CopyIcon />}
                  variant="outline"
                  onClick={() => copyToClipboard(selectedToken.token || '', '访问令牌')}
                >
                  复制令牌
                </Button>
              )}
              <Button
                size="sm"
                leftIcon={<ExternalLinkIcon />}
                colorScheme="blue"
                variant="outline"
                onClick={openInExternalClient}
              >
                打开客户端
              </Button>
            </HStack>
          </VStack>

          {/* Instructions */}
          {showInstructions && (
            <Alert status="info" size="sm">
              <AlertIcon />
              <VStack align="start" spacing={1}>
                <Text fontSize="sm" fontWeight="medium">
                  使用外部客户端:
                </Text>
                <Text fontSize="xs">
                  1. 复制上方WebDAV地址到您的文件管理器
                </Text>
                <Text fontSize="xs">
                  2. 输入用户名和访问令牌进行认证
                </Text>
                <Text fontSize="xs">
                  3. 支持: Windows资源管理器、macOS Finder、第三方工具
                </Text>
              </VStack>
            </Alert>
          )}
        </VStack>
      </CardBody>
    </Card>
  );
}
