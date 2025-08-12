'use client';

import { useState, useEffect } from 'react';
import {
  Box,
  VStack,
  HStack,
  Text,
  Button,
  Input,
  Select,
  Card,
  CardBody,
  CardHeader,
  Heading,
  Code,
  Alert,
  AlertIcon,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  IconButton,
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
  Flex,
  Spacer,
  Tooltip,
} from '@chakra-ui/react';
import { WebDAVAPI, WebDAVToken, CreateTokenRequest } from '@/lib/api/webdav';
import { DeleteIcon, CopyIcon, ExternalLinkIcon } from '@chakra-ui/icons';

export default function WebDAVPage() {
  const [buckets, setBuckets] = useState<string[]>([]);
  const [tokens, setTokens] = useState<WebDAVToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedBucket, setSelectedBucket] = useState<string>('');
  const [newToken, setNewToken] = useState<WebDAVToken | null>(null);
  const [expiresIn, setExpiresIn] = useState<number>(24);
  
  const { isOpen, onOpen, onClose } = useDisclosure();
  const toast = useToast();

  // Load buckets and tokens on component mount
  useEffect(() => {
    loadBuckets();
    loadTokens();
  }, []);

  const loadBuckets = async () => {
    try {
      const bucketList = await WebDAVAPI.getAccessibleBuckets();
      setBuckets(bucketList);
    } catch (error) {
      toast({
        title: '加载存储桶失败',
        description: '无法获取可访问的存储桶列表',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const loadTokens = async () => {
    try {
      const tokenList = await WebDAVAPI.listTokens();
      setTokens(tokenList);
    } catch (error) {
      toast({
        title: '加载令牌失败',
        description: '无法获取WebDAV令牌列表',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  const handleCreateToken = async () => {
    if (!selectedBucket) {
      toast({
        title: '请选择存储桶',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setLoading(true);
    try {
      const request: CreateTokenRequest = {
        bucket: selectedBucket,
        expires_in: expiresIn,
      };
      const token = await WebDAVAPI.createToken(request);
      setNewToken(token);
      loadTokens(); // Refresh token list
      toast({
        title: '令牌创建成功',
        description: '请复制并保存令牌，它将不会再次显示',
        status: 'success',
        duration: 5000,
        isClosable: true,
      });
    } catch (error) {
      toast({
        title: '创建令牌失败',
        description: '无法创建WebDAV访问令牌',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteToken = async (tokenId: number) => {
    if (!confirm('确定要删除这个令牌吗？')) return;

    try {
      await WebDAVAPI.deleteToken(tokenId);
      loadTokens(); // Refresh token list
      toast({
        title: '令牌删除成功',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error) {
      toast({
        title: '删除令牌失败',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const isExpired = (dateString: string) => {
    return new Date(dateString) < new Date();
  };

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        <Flex align="center">
          <Heading size="lg">WebDAV 访问管理</Heading>
          <Spacer />
          <Button
            as="a"
            href="https://github.com/myysophia/ossmanager/blob/main/docs/webdav-usage.md"
            target="_blank"
            rel="noopener noreferrer"
            leftIcon={<ExternalLinkIcon />}
            variant="outline"
            mr={3}
          >
            📄 使用指南
          </Button>
          <Button colorScheme="blue" onClick={onOpen}>
            创建新令牌
          </Button>
        </Flex>
        
        {/* Help Info Card */}
        <Alert status="info" borderRadius="lg">
          <AlertIcon />
          <Box>
            <Text fontWeight="bold" mb={1}>需要帮助？</Text>
            <Text fontSize="sm" mb={2}>
              查看完整的 WebDAV 使用指南，包含详细的挂载示例、认证配置和常见问题解决方案。
            </Text>
            <Button
              as="a"
              href="https://github.com/myysophia/ossmanager/blob/main/docs/webdav-usage.md"
              target="_blank"
              rel="noopener noreferrer"
              size="sm"
              leftIcon={<ExternalLinkIcon />}
              colorScheme="blue"
              variant="outline"
            >
              📖 阅读完整指南
            </Button>
          </Box>
        </Alert>
        
        {/* Token List */}
        <Card>
          <CardHeader>
            <Heading size="md">访问令牌列表</Heading>
          </CardHeader>
          <CardBody>
            {tokens.length === 0 ? (
              <Text color="gray.500" textAlign="center" py={8}>
                暂无WebDAV访问令牌
              </Text>
            ) : (
              <Table variant="simple">
                <Thead>
                  <Tr>
                    <Th>存储桶</Th>
                    <Th>WebDAV地址</Th>
                    <Th>创建时间</Th>
                    <Th>过期时间</Th>
                    <Th>状态</Th>
                    <Th>操作</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {tokens.map((token) => (
                    <Tr key={token.id}>
                      <Td>
                        <Code>{token.bucket}</Code>
                      </Td>
                      <Td>
                        <HStack>
                          <Code fontSize="sm">{getWebDAVUrl(token.bucket)}</Code>
                          <Tooltip label="复制地址">
                            <IconButton
                              aria-label="Copy URL"
                              icon={<CopyIcon />}
                              size="sm"
                              variant="ghost"
                              onClick={() => copyToClipboard(getWebDAVUrl(token.bucket))}
                            />
                          </Tooltip>
                        </HStack>
                      </Td>
                      <Td>{formatDate(token.created_at)}</Td>
                      <Td>{formatDate(token.expires_at)}</Td>
                      <Td>
                        <Badge
                          colorScheme={isExpired(token.expires_at) ? 'red' : 'green'}
                        >
                          {isExpired(token.expires_at) ? '已过期' : '有效'}
                        </Badge>
                      </Td>
                      <Td>
                        <Tooltip label="删除令牌">
                          <IconButton
                            aria-label="Delete token"
                            icon={<DeleteIcon />}
                            size="sm"
                            colorScheme="red"
                            variant="ghost"
                            onClick={() => handleDeleteToken(token.id)}
                          />
                        </Tooltip>
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
            )}
          </CardBody>
        </Card>

        {/* Token Creation Modal */}
        <Modal isOpen={isOpen} onClose={onClose} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>创建WebDAV访问令牌</ModalHeader>
            <ModalCloseButton />
            <ModalBody>
              <VStack spacing={4}>
                <FormControl>
                  <FormLabel>存储桶</FormLabel>
                  <Select
                    placeholder="选择存储桶"
                    value={selectedBucket}
                    onChange={(e) => setSelectedBucket(e.target.value)}
                  >
                    {buckets.map((bucket) => (
                      <option key={bucket} value={bucket}>
                        {bucket}
                      </option>
                    ))}
                  </Select>
                </FormControl>
                
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
                        <Code p={2} fontSize="sm" wordBreak="break-all">
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
              <Button mr={3} onClick={onClose}>
                关闭
              </Button>
              <Button
                colorScheme="blue"
                onClick={handleCreateToken}
                isLoading={loading}
                isDisabled={!selectedBucket}
              >
                创建令牌
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>

        <Tabs>
          <TabList>
            <Tab>Windows 配置</Tab>
            <Tab>macOS 配置</Tab>
            <Tab>Linux 配置</Tab>
            <Tab>第三方客户端</Tab>
          </TabList>

          <TabPanels>
            <TabPanel>
              <Card>
                <CardBody>
                  <VStack align="stretch" spacing={4}>
                    <Heading size="sm">Windows 文件资源管理器配置</Heading>
                    <Text>1. 打开文件资源管理器</Text>
                    <Text>2. 右键"此电脑" → "映射网络驱动器"</Text>
                    <Text>3. 选择驱动器号，输入地址（如：http://your-domain/webdav/bucket-name）</Text>
                    <Text>4. 输入您的用户名</Text>
                    <Text>5. 输入WebDAV访问令牌作为密码</Text>
                    <Alert status="info" mt={4}>
                      <AlertIcon />
                      <Box>
                        <Text fontSize="sm" mb={2}>
                          <strong>认证方式：</strong>可以使用用户名/令牌，或者直接使用 Token xxx 作为认证头
                        </Text>
                        <Text fontSize="xs" color="blue.600">
                          📄 更多 Windows 配置细节和故障排查，请查看
                          <Button
                            as="a"
                            href="https://github.com/myysophia/ossmanager/blob/main/docs/webdav-usage.md#windows-%E7%B3%BB%E7%BB%9F"
                            target="_blank"
                            rel="noopener noreferrer"
                            size="xs"
                            variant="link"
                            colorScheme="blue"
                            ml={1}
                          >
                            完整指南
                          </Button>
                        </Text>
                      </Box>
                    </Alert>
                  </VStack>
                </CardBody>
              </Card>
            </TabPanel>
            
            <TabPanel>
              <Card>
                <CardBody>
                  <VStack align="stretch" spacing={4}>
                    <Heading size="sm">macOS Finder 配置</Heading>
                    <Text>1. 打开 Finder</Text>
                    <Text>2. 按 Cmd+K 或菜单"前往" → "连接服务器"</Text>
                    <Text>3. 输入服务器地址（如：http://your-domain/webdav/bucket-name）</Text>
                    <Text>4. 输入您的用户名和WebDAV访问令牌</Text>
                  </VStack>
                </CardBody>
              </Card>
            </TabPanel>
            
            <TabPanel>
              <Card>
                <CardBody>
                  <VStack align="stretch" spacing={4}>
                    <Heading size="sm">Linux 命令行配置</Heading>
                    <Text>使用 davfs2：</Text>
                    <Code>
                      sudo mount -t davfs http://your-domain/webdav/bucket-name /mnt/webdav
                    </Code>
                    <Text>或使用 cadaver：</Text>
                    <Code>
                      cadaver http://your-domain/webdav/bucket-name
                    </Code>
                    <Alert status="info" mt={4}>
                      <AlertIcon />
                      <Box>
                        <Text fontSize="sm">
                          <strong>提示：</strong>在连接时输入您的用户名和 WebDAV 访问令牌
                        </Text>
                      </Box>
                    </Alert>
                  </VStack>
                </CardBody>
              </Card>
            </TabPanel>
            
            <TabPanel>
              <Card>
                <CardBody>
                  <VStack align="stretch" spacing={4}>
                    <Heading size="sm">推荐的第三方客户端</Heading>
                    <Text>• WinSCP (Windows)</Text>
                    <Text>• Cyberduck (跨平台)</Text>
                    <Text>• FileZilla (跨平台)</Text>
                    <Text>• WebDAV Navigator (iOS/Android)</Text>
                  </VStack>
                </CardBody>
              </Card>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>
    </Box>
  );
}

