'use client';

import { useState } from 'react';
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
} from '@chakra-ui/react';
import { WebDAVAPI } from '@/lib/api/webdav';

export default function WebDAVPage() {
  const [buckets, setBuckets] = useState<string[]>(['default-bucket']);
  const [selectedBucket, setSelectedBucket] = useState<string>('');
  const [accessToken, setAccessToken] = useState<string>('');

  const handleGenerateToken = async () => {
    if (!selectedBucket) return;
    
    // Mock API call
    setAccessToken('sample-access-token');
  };

  const webdavUrl = `${window.location.origin}/webdav/${selectedBucket}`;

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        <Heading size="lg">WebDAV 访问配置</Heading>
        
        <Card>
          <CardHeader>
            <Heading size="md">基本配置</Heading>
          </CardHeader>
          <CardBody>
            <VStack spacing={4} align="stretch">
              <Box>
                <Text mb={2}>选择存储桶：</Text>
                <Select
                  placeholder="选择要访问的存储桶"
                  value={selectedBucket}
                  onChange={(e) => setSelectedBucket(e.target.value)}
                >
                  {buckets.map((bucket) => (
                    <option key={bucket} value={bucket}>
                      {bucket}
                    </option>
                  ))}
                </Select>
              </Box>
              
              {selectedBucket && (
                <Box>
                  <Text mb={2}>WebDAV 地址：</Text>
                  <Code p={2} borderRadius="md" bg="gray.100">
                    {webdavUrl}
                  </Code>
                </Box>
              )}
              
              <Button
                colorScheme="blue"
                onClick={handleGenerateToken}
                isDisabled={!selectedBucket}
              >
                生成访问令牌
              </Button>
              
              {accessToken && (
                <Box>
                  <Text mb={2}>访问令牌：</Text>
                  <Code p={2} borderRadius="md" bg="gray.100">
                    {accessToken}
                  </Code>
                </Box>
              )}
            </VStack>
          </CardBody>
        </Card>

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
                    <Text>3. 选择驱动器号，输入地址：</Text>
                    <Code>{webdavUrl}</Code>
                    <Text>4. 输入用户名和访问令牌作为密码</Text>
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
                    <Text>3. 输入服务器地址：</Text>
                    <Code>{webdavUrl}</Code>
                    <Text>4. 输入用户名和访问令牌</Text>
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
                      sudo mount -t davfs {webdavUrl} /mnt/webdav
                    </Code>
                    <Text>或使用 cadaver：</Text>
                    <Code>
                      cadaver {webdavUrl}
                    </Code>
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

