'use client';

import React, { useState, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Select,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Code,
  Alert,
  AlertIcon,
  useToast,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  FormControl,
  FormLabel,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Badge,
  IconButton,
  Tooltip,
  Spinner,
  Flex,
  Spacer,
} from '@chakra-ui/react';
import { WebDAVAPI, WebDAVToken, CreateTokenRequest } from '@/lib/api/webdav';
import { CopyIcon, AddIcon, ExternalLinkIcon } from '@chakra-ui/icons';

export interface TokenPickerProps {
  selectedBucket: string;
  onTokenSelect?: (token: WebDAVToken | null) => void;
  showWebDAVUrl?: boolean;
  showCreateOption?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export default function TokenPicker({
  selectedBucket,
  onTokenSelect,
  showWebDAVUrl = true,
  showCreateOption = true,
  size = 'md'
}: TokenPickerProps) {
  const [tokens, setTokens] = useState<WebDAVToken[]>([]);
  const [selectedTokenId, setSelectedTokenId] = useState<number | null>(null);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [newToken, setNewToken] = useState<WebDAVToken | null>(null);
  const [expiresIn, setExpiresIn] = useState<number>(24);
  
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  // Load tokens when bucket changes
  useEffect(() => {
    if (selectedBucket) {
      loadTokensForBucket();
    } else {
      setTokens([]);
      setSelectedTokenId(null);
    }
  }, [selectedBucket]);

  // Notify parent when selected token changes
  useEffect(() => {
    const selectedToken = tokens.find(token => token.id === selectedTokenId) || null;
    if (onTokenSelect) {
      onTokenSelect(selectedToken);
    }
  }, [selectedTokenId, tokens, onTokenSelect]);

  const loadTokensForBucket = async () => {
    setLoading(true);
    try {
      const tokenList = await WebDAVAPI.listTokens(selectedBucket, false); // Exclude expired tokens
      setTokens(tokenList);
      
      // Auto-select first valid token if available
      if (tokenList.length > 0 && !selectedTokenId) {
        const validToken = tokenList.find(token => !isExpired(token.expires_at));
        if (validToken) {
          setSelectedTokenId(validToken.id);
        }
      }
    } catch (error) {
      console.error('Load tokens error:', error);
      toast({
        title: '加载令牌失败',
        description: '无法获取WebDAV令牌列表',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCreateToken = async () => {
    if (!selectedBucket) return;

    setCreating(true);
    try {
      const request: CreateTokenRequest = {
        bucket: selectedBucket,
        expires_in: expiresIn,
      };
      const token = await WebDAVAPI.createToken(request);
      setNewToken(token);
      
      // Refresh token list and select the new token
      await loadTokensForBucket();
      setSelectedTokenId(token.id);
      
      toast({
        title: '令牌创建成功',
        description: '新的WebDAV访问令牌已创建',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error) {
      console.error('Create token error:', error);
      toast({
        title: '创建令牌失败',
        description: '无法创建WebDAV访问令牌',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setCreating(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({
        title: '已复制到剪贴板',
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

  const isExpired = (dateString: string) => {
    return new Date(dateString) < new Date();
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const selectedToken = tokens.find(token => token.id === selectedTokenId);
  const webdavUrl = selectedBucket ? getWebDAVUrl(selectedBucket) : '';

  if (!selectedBucket) {
    return (
      <Alert status="info">
        <AlertIcon />
        <Text>请先选择存储桶</Text>
      </Alert>
    );
  }

  return (
    <Card size={size}>
      <CardHeader pb={2}>
        <Flex align="center">
          <Heading size="sm">WebDAV 访问配置</Heading>
          <Spacer />
          {showCreateOption && (
            <Tooltip label="创建新令牌">
              <IconButton
                aria-label="创建新令牌"
                icon={<AddIcon />}
                size="sm"
                variant="outline"
                onClick={onOpen}
                isDisabled={!selectedBucket}
              />
            </Tooltip>
          )}
        </Flex>
      </CardHeader>
      <CardBody pt={0}>
        <VStack spacing={4} align="stretch">
          {/* Token Selection */}
          <FormControl>
            <FormLabel fontSize="sm">访问令牌</FormLabel>
            {loading ? (
              <HStack>
                <Spinner size="sm" />
                <Text fontSize="sm">加载令牌...</Text>
              </HStack>
            ) : tokens.length === 0 ? (
              <Alert status="warning" size="sm">
                <AlertIcon />
                <Text fontSize="sm">
                  该存储桶暂无可用的访问令牌
                  {showCreateOption && '，请创建一个新令牌'}
                </Text>
              </Alert>
            ) : (
              <Select
                size="sm"
                placeholder="选择访问令牌"
                value={selectedTokenId || ''}
                onChange={(e) => setSelectedTokenId(e.target.value ? parseInt(e.target.value) : null)}
              >
                {tokens.map((token) => (
                  <option key={token.id} value={token.id}>
                    {token.bucket} - {isExpired(token.expires_at) ? '已过期' : '有效'} (到期: {formatDate(token.expires_at)})
                  </option>
                ))}
              </Select>
            )}
          </FormControl>

          {/* Token Details */}
          {selectedToken && (
            <Box>
              <HStack mb={2}>
                <Text fontSize="sm" fontWeight="medium">令牌状态:</Text>
                <Badge colorScheme={isExpired(selectedToken.expires_at) ? 'red' : 'green'}>
                  {isExpired(selectedToken.expires_at) ? '已过期' : '有效'}
                </Badge>
              </HStack>
              <Text fontSize="xs" color="gray.600">
                创建时间: {formatDate(selectedToken.created_at)}
              </Text>
              <Text fontSize="xs" color="gray.600">
                过期时间: {formatDate(selectedToken.expires_at)}
              </Text>
            </Box>
          )}

          {/* WebDAV URL */}
          {showWebDAVUrl && (
            <FormControl>
              <FormLabel fontSize="sm">WebDAV 地址</FormLabel>
              <HStack>
                <Code fontSize="sm" p={2} borderRadius="md" bg="gray.50" flex={1}>
                  {webdavUrl}
                </Code>
                <Tooltip label="复制地址">
                  <IconButton
                    aria-label="复制WebDAV地址"
                    icon={<CopyIcon />}
                    size="sm"
                    variant="ghost"
                    onClick={() => copyToClipboard(webdavUrl)}
                  />
                </Tooltip>
                <Tooltip label="在外部WebDAV客户端中打开">
                  <IconButton
                    aria-label="打开WebDAV"
                    icon={<ExternalLinkIcon />}
                    size="sm"
                    variant="ghost"
                    onClick={() => window.open(webdavUrl, '_blank')}
                  />
                </Tooltip>
              </HStack>
              <Text fontSize="xs" color="gray.600" mt={1}>
                复制此地址到您的操作系统文件管理器或WebDAV客户端中挂载
              </Text>
            </FormControl>
          )}

          {/* Usage Info */}
          <Alert status="info" size="sm">
            <AlertIcon />
            <VStack align="start" spacing={1}>
              <Text fontSize="sm" fontWeight="medium">使用说明</Text>
              <Text fontSize="xs">
                • 浏览器内使用代理API无需令牌（自动JWT认证）
              </Text>
              <Text fontSize="xs">
                • 外部客户端需要令牌进行身份验证
              </Text>
              <Text fontSize="xs">
                • 用户名: 您的登录用户名，密码: 访问令牌
              </Text>
            </VStack>
          </Alert>
        </VStack>
      </CardBody>

      {/* Create Token Modal */}
      <Modal isOpen={isOpen} onClose={() => {
        onClose();
        setNewToken(null);
      }} size="lg">
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>创建WebDAV访问令牌</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <VStack spacing={4}>
              <Alert status="info" size="sm">
                <AlertIcon />
                <Text fontSize="sm">
                  为存储桶 <Code>{selectedBucket}</Code> 创建新的访问令牌
                </Text>
              </Alert>

              <FormControl>
                <FormLabel>过期时间（小时）</FormLabel>
                <NumberInput
                  value={expiresIn}
                  onChange={(_, value) => setExpiresIn(value || 24)}
                  min={1}
                  max={720} // Max 30 days
                >
                  <NumberInputField />
                  <NumberInputStepper>
                    <NumberIncrementStepper />
                    <NumberDecrementStepper />
                  </NumberInputStepper>
                </NumberInput>
                <Text fontSize="sm" color="gray.500" mt={1}>
                  最大720小时（30天）
                </Text>
              </FormControl>

              {newToken && (
                <Alert status="success">
                  <AlertIcon />
                  <Box>
                    <Text fontWeight="bold" mb={2}>令牌创建成功！</Text>
                    <Text fontSize="sm" mb={2}>请复制并保存以下令牌，它将不会再次显示：</Text>
                    <HStack>
                      <Code p={2} fontSize="sm" wordBreak="break-all" flex={1}>
                        {newToken.token}
                      </Code>
                      <Button
                        size="sm"
                        leftIcon={<CopyIcon />}
                        onClick={() => copyToClipboard(newToken.token || '')}
                      >
                        复制
                      </Button>
                    </HStack>
                  </Box>
                </Alert>
              )}
            </VStack>
          </ModalBody>
          <ModalFooter>
            <Button mr={3} onClick={() => {
              onClose();
              setNewToken(null);
            }}>
              关闭
            </Button>
            <Button
              colorScheme="blue"
              onClick={handleCreateToken}
              isLoading={creating}
              isDisabled={!selectedBucket || !!newToken}
            >
              创建令牌
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Card>
  );
}
