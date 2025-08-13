'use client';

import React, { useState, useEffect, useCallback } from 'react';
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
  Alert,
  AlertIcon,
  Spinner,
  useToast,
  Badge,
  Flex,
  Spacer,
  IconButton,
  Tooltip,
  Divider,
  Grid,
  GridItem,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react';
import { WebDAVAPI, WebDAVToken } from '@/lib/api/webdav';
import { FiRefreshCw, FiFolder, FiCloud, FiExternalLink, FiSettings, FiFile } from 'react-icons/fi';
import TokenPicker from '@/components/webdav/TokenPicker';
import WebDAVUrlHelper from '@/components/webdav/WebDAVUrlHelper';
import { FileExplorer } from '@/components/file-explorer';
import { WebDAVBrowserAPI } from '@/lib/api/webdavBrowser';
import type { FileSystemItem, BreadcrumbItem, UploadProgress, FileExplorerActions } from '@/types/file-explorer';

export default function WebDAVBrowserPage() {
  const [buckets, setBuckets] = useState<string[]>([]);
  const [selectedBucket, setSelectedBucket] = useState<string>('');
  const [selectedToken, setSelectedToken] = useState<WebDAVToken | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  
  // File explorer state
  const [currentPath, setCurrentPath] = useState<string>('/');
  const [items, setItems] = useState<FileSystemItem[]>([]);
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([]);
  const [uploads, setUploads] = useState<UploadProgress[]>([]);
  const [isFilesLoading, setIsFilesLoading] = useState(false);
  const [isFilesRefreshing, setIsFilesRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState<'files' | 'settings'>('files');
  
  const toast = useToast();

  // Load accessible buckets on component mount
  useEffect(() => {
    loadBuckets();
  }, []);
  
  // Load files when bucket changes
  useEffect(() => {
    if (selectedBucket) {
      loadFiles(currentPath);
    }
  }, [selectedBucket]);
  
  // Update breadcrumbs when path changes
  useEffect(() => {
    updateBreadcrumbs(currentPath);
  }, [currentPath]);

  const loadBuckets = async () => {
    try {
      setLoading(true);
      const bucketList = await WebDAVAPI.getAccessibleBuckets();
      setBuckets(bucketList);
      
      // Auto-select first bucket if available
      if (bucketList.length > 0 && !selectedBucket) {
        setSelectedBucket(bucketList[0]);
      }
    } catch (error) {
      console.error('Load buckets error:', error);
      toast({
        title: '加载存储桶失败',
        description: '无法获取可访问的存储桶列表',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadBuckets();
    setRefreshing(false);
  };

  const getWebDAVUrl = (bucket: string) => {
    if (typeof window !== 'undefined') {
      return `${window.location.origin}/webdav/${bucket}`;
    }
    return `/webdav/${bucket}`;
  };

  const openWebDAVUrl = (bucket: string) => {
    const url = getWebDAVUrl(bucket);
    window.open(url, '_blank');
  };
  
  // File explorer functions
  const loadFiles = useCallback(async (path: string) => {
    if (!selectedBucket) return;
    
    try {
      setIsFilesLoading(true);
      const response = await WebDAVBrowserAPI.listObjects(selectedBucket, path);
      
      const fileItems: FileSystemItem[] = response.items.map(obj => ({
        id: obj.path,
        name: obj.name,
        type: obj.isDir ? 'folder' : 'file',
        size: obj.size,
        lastModified: new Date(obj.mtime),
        path: obj.path,
        permissions: ['read', 'write', 'delete']
      }));
      
      setItems(fileItems);
    } catch (error) {
      console.error('Load files error:', error);
      toast({
        title: '加载文件失败',
        description: '无法获取文件列表',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsFilesLoading(false);
    }
  }, [selectedBucket, toast]);
  
  const updateBreadcrumbs = useCallback((path: string) => {
    const pathParts = path.split('/').filter(Boolean);
    const crumbs: BreadcrumbItem[] = [
      { label: '根目录', path: '/' }
    ];
    
    let currentPath = '';
    pathParts.forEach(part => {
      currentPath += '/' + part;
      crumbs.push({
        label: part,
        path: currentPath
      });
    });
    
    setBreadcrumbs(crumbs);
  }, []);
  
  const navigateToPath = useCallback(async (path: string) => {
    setCurrentPath(path);
    await loadFiles(path);
  }, [loadFiles]);
  
  // File explorer actions
  const fileActions: FileExplorerActions = {
    onFolderDoubleClick: async (item) => {
      await navigateToPath(item.path);
    },
    
    onFileDoubleClick: async (item) => {
      // Download file
      try {
        const url = `/api/v1/webdav/objects/${selectedBucket}?path=${encodeURIComponent(item.path)}`;
        window.open(url, '_blank');
      } catch (error) {
        toast({
          title: '下载失败',
          description: '无法下载文件',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
      }
    },
    
    onUpload: async (files, path) => {
      const uploadPromises = files.map(async (file) => {
        const uploadPath = path === '/' ? file.name : `${path}/${file.name}`;
        
        try {
          await WebDAVBrowserAPI.uploadFile(selectedBucket, uploadPath, file, {
            onProgress: (progress) => {
              setUploads(prev => {
                const existing = prev.find(u => u.id === file.name);
                if (existing) {
                  return prev.map(u => u.id === file.name ? { ...u, progress } : u);
                } else {
                  return [...prev, {
                    id: file.name,
                    filename: file.name,
                    progress,
                    status: 'uploading',
                    size: file.size
                  }];
                }
              });
            }
          });
          
          setUploads(prev => prev.map(u => 
            u.id === file.name ? { ...u, status: 'completed', progress: 100 } : u
          ));
          
          toast({
            title: '上传成功',
            description: `文件 ${file.name} 上传完成`,
            status: 'success',
            duration: 3000,
            isClosable: true,
          });
        } catch (error) {
          setUploads(prev => prev.map(u => 
            u.id === file.name ? { ...u, status: 'failed' } : u
          ));
          
          toast({
            title: '上传失败',
            description: `文件 ${file.name} 上传失败`,
            status: 'error',
            duration: 3000,
            isClosable: true,
          });
        }
      });
      
      await Promise.all(uploadPromises);
      await loadFiles(currentPath);
    },
    
    onDelete: async (items) => {
      try {
        for (const item of items) {
          await WebDAVBrowserAPI.deleteObject(selectedBucket, item.path);
        }
        
        toast({
          title: '删除成功',
          description: `已删除 ${items.length} 个项目`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        
        await loadFiles(currentPath);
      } catch (error) {
        toast({
          title: '删除失败',
          description: '删除操作失败',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
      }
    },
    
    onRename: async (item, newName) => {
      try {
        const newPath = item.path.replace(item.name, newName);
        await WebDAVBrowserAPI.renameObject(selectedBucket, item.path, newPath);
        
        toast({
          title: '重命名成功',
          description: `已重命名为 ${newName}`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        
        await loadFiles(currentPath);
      } catch (error) {
        toast({
          title: '重命名失败',
          description: '重命名操作失败',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
      }
    },
    
    onCreateFolder: async (name, path) => {
      try {
        // 确保路径始终以斜杠开头，符合后端验证要求
        const folderPath = path === '/' ? `/${name}` : `${path}/${name}`;
        await WebDAVBrowserAPI.createFolder(selectedBucket, folderPath);
        
        toast({
          title: '创建成功',
          description: `文件夹 ${name} 创建完成`,
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
        
        await loadFiles(currentPath);
      } catch (error) {
        toast({
          title: '创建失败',
          description: '文件夹创建失败',
          status: 'error',
          duration: 3000,
          isClosable: true,
        });
      }
    },
    
    onRefresh: async () => {
      setIsFilesRefreshing(true);
      await loadFiles(currentPath);
      setIsFilesRefreshing(false);
    },
    
    onBreadcrumbClick: navigateToPath
  };

  if (loading) {
    return (
      <Box p={6}>
        <VStack spacing={4}>
          <Spinner size="lg" />
          <Text>加载存储桶列表...</Text>
        </VStack>
      </Box>
    );
  }

  return (
    <Box p={6}>
      <VStack spacing={6} align="stretch">
        {/* Page Header */}
        <Flex align="center">
          <HStack spacing={3}>
            <Box 
              bg="blue.500" 
              p={2} 
              borderRadius="md" 
              color="white"
            >
              <FiFolder size={20} />
            </Box>
            <Heading size="lg">WebDAV 浏览器</Heading>
          </HStack>
          <Spacer />
          <Tooltip label="刷新存储桶列表">
            <IconButton
              aria-label="刷新"
              icon={<FiRefreshCw />}
              variant="outline"
              isLoading={refreshing}
              onClick={handleRefresh}
            />
          </Tooltip>
        </Flex>

        {/* Bucket Selection Card */}
        <Card>
          <CardHeader>
            <Heading size="md">
              <HStack>
                <FiCloud />
                <Text>存储桶选择</Text>
              </HStack>
            </Heading>
          </CardHeader>
          <CardBody>
            {buckets.length === 0 ? (
              <Alert status="warning">
                <AlertIcon />
                <VStack align="start" spacing={2}>
                  <Text>暂无可访问的存储桶</Text>
                  <Text fontSize="sm" color="gray.600">
                    请联系管理员为您分配存储桶访问权限
                  </Text>
                </VStack>
              </Alert>
            ) : (
              <VStack spacing={4} align="stretch">
                <HStack>
                  <Text fontSize="sm" color="gray.600" minW="80px">
                    当前存储桶:
                  </Text>
                  <Select 
                    value={selectedBucket} 
                    onChange={(e) => setSelectedBucket(e.target.value)}
                    placeholder="请选择存储桶"
                  >
                    {buckets.map((bucket) => (
                      <option key={bucket} value={bucket}>
                        {bucket}
                      </option>
                    ))}
                  </Select>
                  <Badge colorScheme="blue" fontSize="xs">
                    共 {buckets.length} 个
                  </Badge>
                </HStack>

                {selectedBucket && (
                  <>
                    <Divider />
                    <VStack spacing={3} align="stretch">
                      <Text fontSize="sm" fontWeight="medium">
                        WebDAV 连接信息:
                      </Text>
                      <Box 
                        bg="gray.50" 
                        p={3} 
                        borderRadius="md" 
                        border="1px solid" 
                        borderColor="gray.200"
                      >
                        <VStack spacing={2} align="stretch">
                          <HStack>
                            <Text fontSize="sm" color="gray.600" minW="60px">
                              URL:
                            </Text>
                            <Text fontSize="sm" fontFamily="mono" flex={1}>
                              {getWebDAVUrl(selectedBucket)}
                            </Text>
                            <Tooltip label="在新窗口中打开">
                              <IconButton
                                aria-label="打开WebDAV"
                                icon={<FiExternalLink />}
                                size="xs"
                                variant="ghost"
                                onClick={() => openWebDAVUrl(selectedBucket)}
                              />
                            </Tooltip>
                          </HStack>
                          <HStack>
                            <Text fontSize="sm" color="gray.600" minW="60px">
                              存储桶:
                            </Text>
                            <Text fontSize="sm" fontFamily="mono" flex={1}>
                              {selectedBucket}
                            </Text>
                          </HStack>
                        </VStack>
                      </Box>
                    </VStack>
                  </>
                )}
              </VStack>
            )}
          </CardBody>
        </Card>

        {/* File Explorer with Tabs */}
        {selectedBucket && (
          <Card>
            <CardBody p={0}>
              <Tabs 
                index={activeTab === 'files' ? 0 : 1} 
                onChange={(index) => setActiveTab(index === 0 ? 'files' : 'settings')}
                isFitted
                variant="enclosed"
              >
                <TabList>
                  <Tab>
                    <HStack spacing={2}>
                      <FiFile />
                      <Text>文件浏览</Text>
                    </HStack>
                  </Tab>
                  <Tab>
                    <HStack spacing={2}>
                      <FiSettings />
                      <Text>连接设置</Text>
                    </HStack>
                  </Tab>
                </TabList>

                <TabPanels>
                  <TabPanel p={0}>
                    {/* File Explorer */}
                    <Box minH="600px">
                      <FileExplorer
                        currentPath={currentPath}
                        items={items}
                        breadcrumbs={breadcrumbs}
                        uploads={uploads}
                        isLoading={isFilesLoading}
                        isRefreshing={isFilesRefreshing}
                        viewMode="list"
                        actions={fileActions}
                        config={{
                          allowMultiSelect: true,
                          showHiddenFiles: false,
                          maxFileSize: 100 * 1024 * 1024, // 100MB
                          maxConcurrentUploads: 5,
                          enableDragAndDrop: true,
                          enableContextMenu: true,
                        }}
                      />
                    </Box>
                  </TabPanel>
                  
                  <TabPanel p={6}>
                    {/* Connection Settings */}
                    <VStack spacing={6} align="stretch">
                      <Grid templateColumns={{ base: '1fr', lg: '1fr 1fr' }} gap={6}>
                        <GridItem>
                          <TokenPicker
                            selectedBucket={selectedBucket}
                            onTokenSelect={setSelectedToken}
                            showWebDAVUrl={false}
                            size="md"
                          />
                        </GridItem>
                        <GridItem>
                          <WebDAVUrlHelper
                            selectedBucket={selectedBucket}
                            selectedToken={selectedToken}
                            size="md"
                            showInstructions={true}
                          />
                        </GridItem>
                      </Grid>
                      
                      {/* Status and Instructions */}
                      <VStack spacing={3} align="stretch">
                        <Alert status="info">
                          <AlertIcon />
                          <VStack align="start" spacing={1}>
                            <Text fontWeight="medium">WebDAV 浏览器功能</Text>
                            <Text fontSize="sm">
                              • 浏览器内访问: 使用代理API，无需令牌 (自动JWT认证)
                            </Text>
                            <Text fontSize="sm">
                              • 外部客户端: 需要访问令牌进行身份验证
                            </Text>
                            <Text fontSize="sm">
                              • 支持完整WebDAV操作: 浏览、上传、下载、创建文件夹等
                            </Text>
                          </VStack>
                        </Alert>
                        
                        {selectedBucket && selectedToken && (
                          <Alert status="success" size="sm">
                            <AlertIcon />
                            <Text fontSize="sm">
                              ✅ 当前配置完整，可以通过浏览器或外部客户端访问存储桶
                            </Text>
                          </Alert>
                        )}
                        
                        {selectedBucket && !selectedToken && (
                          <Alert status="warning" size="sm">
                            <AlertIcon />
                            <Text fontSize="sm">
                              ⚠️ 浏览器访问正常，但外部客户端需要访问令牌
                            </Text>
                          </Alert>
                        )}
                      </VStack>
                    </VStack>
                  </TabPanel>
                </TabPanels>
              </Tabs>
            </CardBody>
          </Card>
        )}
      </VStack>
    </Box>
  );
}
