'use client';

import React, { useState, useMemo } from 'react';
import {
  Container,
  Box,
  Heading,
  Button,
  Input,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Checkbox,
  IconButton,
  HStack,
  VStack,
  Text,
  Flex,
  Spinner,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Select,
  Badge,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
  Alert,
  AlertIcon,
  useColorModeValue,
} from '@chakra-ui/react';
import {
  FiDownload,
  FiTrash2,
  FiShare2,
  FiRefreshCw,
  FiSearch,
  FiUpload,
  FiFolder,
  FiFile,
  FiGrid,
  FiList,
} from 'react-icons/fi';
import { useFilesWithSWR } from '../../../lib/hooks/useFilesWithSWR';
import { OSSFile, FileQueryParams } from '../../../lib/api/types';
import { ShareLinkModal } from '../../../components/ShareLinkModal';
import { FileUploadZone } from '../../../components/FileUploadZone';
import { optimisticUpdateManager } from '../../../lib/optimistic-updates';

// Format file size
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

// Format date
const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

export default function EnhancedFileListPage() {
  // State
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<number[]>([]);
  const [sortConfig, setSortConfig] = useState<{
    key: keyof OSSFile | '';
    direction: 'asc' | 'desc';
  }>({
    key: 'created_at',
    direction: 'desc',
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table');
  
  // Modals
  const { isOpen: isDeleteOpen, onOpen: onDeleteOpen, onClose: onDeleteClose } = useDisclosure();
  const [fileToDelete, setFileToDelete] = useState<number | null>(null);
  const [shareFile, setShareFile] = useState<OSSFile | null>(null);
  const [isShareModalOpen, setIsShareModalOpen] = useState(false);

  // Query params for API
  const queryParams: FileQueryParams = {
    page: currentPage,
    page_size: pageSize,
    keyword: searchKeyword,
  };

  // Use SWR hook for file management
  const {
    files,
    total,
    loading,
    error,
    deleteFile,
    deleteFiles,
    downloadFile,
    refreshFiles,
  } = useFilesWithSWR(queryParams);

  // Colors
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');

  // Filtered and sorted files for frontend processing
  const filteredAndSortedFiles = useMemo(() => {
    let result = files;

    // Apply sorting if configured
    if (sortConfig.key) {
      result = [...result].sort((a, b) => {
        const aValue = a[sortConfig.key as keyof OSSFile];
        const bValue = b[sortConfig.key as keyof OSSFile];

        if (typeof aValue === 'string' && typeof bValue === 'string') {
          return sortConfig.direction === 'asc'
            ? aValue.localeCompare(bValue)
            : bValue.localeCompare(aValue);
        } else if (typeof aValue === 'number' && typeof bValue === 'number') {
          return sortConfig.direction === 'asc' ? aValue - bValue : bValue - aValue;
        }
        return 0;
      });
    }

    return result;
  }, [files, sortConfig]);

  /**
   * Handle file download
   */
  const handleDownload = async (fileId: number) => {
    try {
      await downloadFile(fileId);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  /**
   * Handle file deletion with optimistic updates
   */
  const handleDelete = async (fileId: number) => {
    try {
      await deleteFile(fileId);
      onDeleteClose();
    } catch (error) {
      console.error('Delete failed:', error);
    }
  };

  /**
   * Handle batch deletion
   */
  const handleBatchDelete = async () => {
    if (selectedFiles.length === 0) return;

    try {
      await deleteFiles(selectedFiles);
      setSelectedFiles([]);
    } catch (error) {
      console.error('Batch delete failed:', error);
    }
  };

  /**
   * Confirm single file deletion
   */
  const confirmDelete = (fileId: number) => {
    setFileToDelete(fileId);
    onDeleteOpen();
  };

  /**
   * Handle share link
   */
  const handleShareLink = (file: OSSFile) => {
    setShareFile(file);
    setIsShareModalOpen(true);
  };

  /**
   * Close share modal
   */
  const handleCloseShareModal = () => {
    setIsShareModalOpen(false);
    setShareFile(null);
  };

  /**
   * Handle search
   */
  const handleSearch = () => {
    setCurrentPage(1);
    // The search will automatically trigger due to queryParams change
  };

  /**
   * Handle sorting
   */
  const handleSort = (field: keyof OSSFile) => {
    setSortConfig(prev => ({
      key: field,
      direction: prev.key === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
  };

  /**
   * Toggle file selection
   */
  const toggleFileSelection = (fileId: number) => {
    setSelectedFiles(prev =>
      prev.includes(fileId)
        ? prev.filter(id => id !== fileId)
        : [...prev, fileId]
    );
  };

  /**
   * Toggle select all
   */
  const toggleSelectAll = () => {
    setSelectedFiles(prev =>
      prev.length === filteredAndSortedFiles.length
        ? []
        : filteredAndSortedFiles.map(f => f.id)
    );
  };

  /**
   * Handle page change
   */
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  /**
   * Handle page size change
   */
  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
  };

  // Get pending actions for UI feedback
  const pendingActions = optimisticUpdateManager.getPendingActions();

  return (
    <Container maxW="container.xl" py={6}>
      <VStack spacing={6} align="stretch">
        {/* Header */}
        <Flex justify="space-between" align="center">
          <Heading size="lg">文件管理</Heading>
          <HStack spacing={3}>
            <Button
              leftIcon={<FiRefreshCw />}
              onClick={refreshFiles}
              isLoading={loading}
              variant="outline"
            >
              刷新
            </Button>
            <Button
              leftIcon={<FiGrid />}
              onClick={() => setViewMode(viewMode === 'table' ? 'grid' : 'table')}
              variant="outline"
            >
              {viewMode === 'table' ? '网格' : '列表'}
            </Button>
          </HStack>
        </Flex>

        {/* Tabs for Upload and File List */}
        <Tabs variant="enclosed">
          <TabList>
            <Tab>
              <HStack>
                <FiUpload />
                <Text>上传文件</Text>
              </HStack>
            </Tab>
            <Tab>
              <HStack>
                <FiFolder />
                <Text>文件列表</Text>
              </HStack>
            </Tab>
          </TabList>

          <TabPanels>
            {/* Upload Tab */}
            <TabPanel>
              <FileUploadZone
                onUploadComplete={() => {
                  refreshFiles();
                }}
                maxFiles={10}
                maxFileSize={500 * 1024 * 1024} // 500MB
                multiple={true}
              />
            </TabPanel>

            {/* File List Tab */}
            <TabPanel>
              {/* Search and Actions Bar */}
              <VStack spacing={4} align="stretch">
                <Flex justify="space-between" align="center" wrap="wrap" gap={4}>
                  <HStack spacing={3} flex={1} maxW="md">
                    <Input
                      placeholder="搜索文件名..."
                      value={searchKeyword}
                      onChange={(e) => setSearchKeyword(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                    />
                    <Button leftIcon={<FiSearch />} onClick={handleSearch}>
                      搜索
                    </Button>
                  </HStack>

                  {selectedFiles.length > 0 && (
                    <Button
                      colorScheme="red"
                      leftIcon={<FiTrash2 />}
                      onClick={handleBatchDelete}
                      size="sm"
                    >
                      批量删除 ({selectedFiles.length})
                    </Button>
                  )}
                </Flex>

                {/* Pending Actions Alert */}
                {pendingActions.length > 0 && (
                  <Alert status="info">
                    <AlertIcon />
                    <Text>
                      正在处理 {pendingActions.length} 个操作...
                    </Text>
                  </Alert>
                )}

                {/* Error State */}
                {error && (
                  <Alert status="error">
                    <AlertIcon />
                    <Text>加载文件列表失败: {error.message}</Text>
                  </Alert>
                )}

                {/* Loading State */}
                {loading ? (
                  <Flex justify="center" align="center" minH="200px">
                    <Spinner size="xl" />
                  </Flex>
                ) : (
                  <>
                    {/* Files Table */}
                    {filteredAndSortedFiles.length === 0 ? (
                      <VStack py={10}>
                        <Text color="gray.500">暂无文件</Text>
                      </VStack>
                    ) : (
                      <Box overflowX="auto" bg={bgColor} borderRadius="md" border="1px solid" borderColor={borderColor}>
                        <Table variant="simple">
                          <Thead>
                            <Tr>
                              <Th width="50px">
                                <Checkbox
                                  isChecked={
                                    selectedFiles.length === filteredAndSortedFiles.length &&
                                    filteredAndSortedFiles.length > 0
                                  }
                                  onChange={toggleSelectAll}
                                />
                              </Th>
                              <Th cursor="pointer" onClick={() => handleSort('original_filename')}>
                                <Flex align="center">
                                  文件名
                                  {sortConfig.key === 'original_filename' && (
                                    <Text ml={1} fontSize="xs">
                                      {sortConfig.direction === 'asc' ? '↑' : '↓'}
                                    </Text>
                                  )}
                                </Flex>
                              </Th>
                              <Th cursor="pointer" onClick={() => handleSort('file_size')}>
                                <Flex align="center">
                                  大小
                                  {sortConfig.key === 'file_size' && (
                                    <Text ml={1} fontSize="xs">
                                      {sortConfig.direction === 'asc' ? '↑' : '↓'}
                                    </Text>
                                  )}
                                </Flex>
                              </Th>
                              <Th>存储路径</Th>
                              <Th cursor="pointer" onClick={() => handleSort('created_at')}>
                                <Flex align="center">
                                  上传时间
                                  {sortConfig.key === 'created_at' && (
                                    <Text ml={1} fontSize="xs">
                                      {sortConfig.direction === 'asc' ? '↑' : '↓'}
                                    </Text>
                                  )}
                                </Flex>
                              </Th>
                              <Th width="150px">操作</Th>
                            </Tr>
                          </Thead>
                          <Tbody>
                            {filteredAndSortedFiles.map((file) => (
                              <Tr key={file.id}>
                                <Td>
                                  <Checkbox
                                    isChecked={selectedFiles.includes(file.id)}
                                    onChange={() => toggleFileSelection(file.id)}
                                  />
                                </Td>
                                <Td>
                                  <HStack>
                                    <FiFile />
                                    <Text noOfLines={1} title={decodeURIComponent(file.original_filename)}>
                                      {decodeURIComponent(file.original_filename)}
                                    </Text>
                                  </HStack>
                                </Td>
                                <Td>{formatFileSize(file.file_size)}</Td>
                                <Td>
                                  <Text noOfLines={1} fontSize="sm" color="gray.600">
                                    {file.object_key || '-'}
                                  </Text>
                                </Td>
                                <Td>{formatDate(file.created_at)}</Td>
                                <Td>
                                  <HStack spacing={2}>
                                    <IconButton
                                      aria-label="下载文件"
                                      icon={<FiDownload />}
                                      size="sm"
                                      variant="ghost"
                                      onClick={() => handleDownload(file.id)}
                                      isLoading={optimisticUpdateManager.isPending('download', file.id)}
                                    />
                                    <IconButton
                                      aria-label="分享链接"
                                      icon={<FiShare2 />}
                                      size="sm"
                                      variant="ghost"
                                      colorScheme="blue"
                                      onClick={() => handleShareLink(file)}
                                    />
                                    <IconButton
                                      aria-label="删除文件"
                                      icon={<FiTrash2 />}
                                      size="sm"
                                      variant="ghost"
                                      colorScheme="red"
                                      onClick={() => confirmDelete(file.id)}
                                      isLoading={optimisticUpdateManager.isPending('delete', file.id)}
                                    />
                                  </HStack>
                                </Td>
                              </Tr>
                            ))}
                          </Tbody>
                        </Table>
                      </Box>
                    )}

                    {/* Pagination */}
                    {filteredAndSortedFiles.length > 0 && (
                      <Flex justify="space-between" align="center">
                        <HStack>
                          <Text fontSize="sm">
                            显示 {Math.min((currentPage - 1) * pageSize + 1, total)} 到{' '}
                            {Math.min(currentPage * pageSize, total)} 条，共 {total} 条记录
                          </Text>
                          <Select
                            value={pageSize}
                            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
                            size="sm"
                            width="auto"
                          >
                            <option value={10}>10条/页</option>
                            <option value={20}>20条/页</option>
                            <option value={50}>50条/页</option>
                            <option value={100}>100条/页</option>
                          </Select>
                        </HStack>
                        
                        <HStack>
                          <Button
                            size="sm"
                            onClick={() => handlePageChange(currentPage - 1)}
                            isDisabled={currentPage === 1 || loading}
                          >
                            上一页
                          </Button>
                          <Text fontSize="sm">
                            第 {currentPage} 页
                          </Text>
                          <Button
                            size="sm"
                            onClick={() => handlePageChange(currentPage + 1)}
                            isDisabled={currentPage * pageSize >= total || loading}
                          >
                            下一页
                          </Button>
                        </HStack>
                      </Flex>
                    )}
                  </>
                )}
              </VStack>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </VStack>

      {/* Delete Confirmation Modal */}
      <Modal isOpen={isDeleteOpen} onClose={onDeleteClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>确认删除</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            您确定要删除选中的文件吗？此操作无法撤销。
          </ModalBody>
          <ModalFooter>
            <Button variant="ghost" mr={3} onClick={onDeleteClose}>
              取消
            </Button>
            <Button
              colorScheme="red"
              onClick={() => fileToDelete && handleDelete(fileToDelete)}
              isLoading={fileToDelete ? optimisticUpdateManager.isPending('delete', fileToDelete) : false}
            >
              确认删除
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Share Link Modal */}
      <ShareLinkModal
        isOpen={isShareModalOpen}
        onClose={handleCloseShareModal}
        file={shareFile}
      />
    </Container>
  );
}
