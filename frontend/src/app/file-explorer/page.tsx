'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Container,
  Heading,
  VStack,
  useToast,
} from '@chakra-ui/react';
import { FileExplorer } from '../../components/file-explorer';
import {
  FileSystemItem,
  BreadcrumbItem,
  UploadProgress,
  FileExplorerActions,
} from '../../components/file-explorer';

// Mock data for demonstration
const generateMockFiles = (path: string): FileSystemItem[] => {
  const baseFiles: Omit<FileSystemItem, 'id' | 'path' | 'parentPath'>[] = [
    {
      name: 'Documents',
      type: 'folder' as const,
      size: 0,
      lastModified: new Date('2024-01-15'),
      filename: 'Documents',
      originalFilename: 'Documents',
      fileSize: 0,
      mimeType: 'application/x-directory',
      md5: '',
      storageType: 'LOCAL' as any,
      bucket: '',
      objectKey: '',
      status: 'COMPLETE' as any,
      uploaderId: 'user1',
      createdAt: '2024-01-15T00:00:00Z',
      updatedAt: '2024-01-15T00:00:00Z',
      isPublic: false,
    },
    {
      name: 'Images',
      type: 'folder' as const,
      size: 0,
      lastModified: new Date('2024-01-10'),
      filename: 'Images',
      originalFilename: 'Images',
      fileSize: 0,
      mimeType: 'application/x-directory',
      md5: '',
      storageType: 'LOCAL' as any,
      bucket: '',
      objectKey: '',
      status: 'COMPLETE' as any,
      uploaderId: 'user1',
      createdAt: '2024-01-10T00:00:00Z',
      updatedAt: '2024-01-10T00:00:00Z',
      isPublic: false,
    },
    {
      name: 'report.pdf',
      type: 'file' as const,
      size: 2048576,
      lastModified: new Date('2024-01-20'),
      filename: 'report.pdf',
      originalFilename: 'report.pdf',
      fileSize: 2048576,
      mimeType: 'application/pdf',
      md5: 'abc123',
      storageType: 'LOCAL' as any,
      bucket: '',
      objectKey: '',
      status: 'COMPLETE' as any,
      uploaderId: 'user1',
      createdAt: '2024-01-20T00:00:00Z',
      updatedAt: '2024-01-20T00:00:00Z',
      isPublic: false,
    },
    {
      name: 'presentation.pptx',
      type: 'file' as const,
      size: 5242880,
      lastModified: new Date('2024-01-18'),
      filename: 'presentation.pptx',
      originalFilename: 'presentation.pptx',
      fileSize: 5242880,
      mimeType: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
      md5: 'def456',
      storageType: 'LOCAL' as any,
      bucket: '',
      objectKey: '',
      status: 'COMPLETE' as any,
      uploaderId: 'user1',
      createdAt: '2024-01-18T00:00:00Z',
      updatedAt: '2024-01-18T00:00:00Z',
      isPublic: false,
    },
    {
      name: 'photo.jpg',
      type: 'file' as const,
      size: 1024000,
      lastModified: new Date('2024-01-22'),
      filename: 'photo.jpg',
      originalFilename: 'photo.jpg',
      fileSize: 1024000,
      mimeType: 'image/jpeg',
      md5: 'ghi789',
      storageType: 'LOCAL' as any,
      bucket: '',
      objectKey: '',
      status: 'COMPLETE' as any,
      uploaderId: 'user1',
      createdAt: '2024-01-22T00:00:00Z',
      updatedAt: '2024-01-22T00:00:00Z',
      isPublic: false,
    },
  ];

  return baseFiles.map((file, index) => ({
    ...file,
    id: `${path}_${index}`,
    path: path === '/' ? `/${file.name}` : `${path}/${file.name}`,
    parentPath: path,
  }));
};

// Generate breadcrumbs from path
const generateBreadcrumbs = (path: string): BreadcrumbItem[] => {
  if (path === '/') {
    return [{ name: '根目录', path: '/' }];
  }

  const parts = path.split('/').filter(Boolean);
  const breadcrumbs: BreadcrumbItem[] = [
    { name: '根目录', path: '/' },
  ];

  let currentPath = '';
  for (const part of parts) {
    currentPath += `/${part}`;
    breadcrumbs.push({
      name: part,
      path: currentPath,
    });
  }

  return breadcrumbs;
};

export default function FileExplorerPage() {
  const [currentPath, setCurrentPath] = useState('/');
  const [items, setItems] = useState<FileSystemItem[]>([]);
  const [uploads, setUploads] = useState<UploadProgress[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  
  const toast = useToast();

  // Load items for current path
  const loadItems = useCallback(async (path: string) => {
    setIsLoading(true);
    
    // Simulate API call delay
    await new Promise(resolve => setTimeout(resolve, 500));
    
    const mockItems = generateMockFiles(path);
    setItems(mockItems);
    setIsLoading(false);
  }, []);

  // Initial load
  useEffect(() => {
    loadItems(currentPath);
  }, [currentPath, loadItems]);

  // File explorer actions
  const actions: FileExplorerActions = {
    onNavigate: useCallback((path: string) => {
      setCurrentPath(path);
    }, []),

    onFileSelect: useCallback((file: FileSystemItem) => {
      console.log('File selected:', file);
    }, []),

    onFileDoubleClick: useCallback((file: FileSystemItem) => {
      toast({
        title: '文件下载',
        description: `正在下载文件: ${file.name}`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onFolderDoubleClick: useCallback((folder: FileSystemItem) => {
      const newPath = folder.path;
      setCurrentPath(newPath);
    }, []),

    onUpload: useCallback((files: File[], path: string) => {
      console.log('Uploading files:', files, 'to path:', path);
      
      // Create upload progress items
      const newUploads: UploadProgress[] = files.map(file => ({
        id: `upload_${Date.now()}_${file.name}`,
        file,
        progress: 0,
        status: 'uploading',
      }));

      setUploads(prev => [...prev, ...newUploads]);

      // Simulate upload progress
      newUploads.forEach((upload, index) => {
        const interval = setInterval(() => {
          setUploads(prev => 
            prev.map(u => {
              if (u.id === upload.id) {
                const newProgress = Math.min(u.progress + Math.random() * 20, 100);
                return {
                  ...u,
                  progress: newProgress,
                  status: newProgress >= 100 ? 'completed' : 'uploading',
                };
              }
              return u;
            })
          );
        }, 300);

        // Stop simulation after completion
        setTimeout(() => {
          clearInterval(interval);
          setUploads(prev => 
            prev.map(u => u.id === upload.id ? { ...u, progress: 100, status: 'completed' } : u)
          );
        }, 3000 + index * 1000);
      });

      toast({
        title: '开始上传',
        description: `正在上传 ${files.length} 个文件`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onRename: useCallback(async (item: FileSystemItem, newName: string) => {
      console.log('Renaming item:', item.name, 'to:', newName);
      
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Update local state
      setItems(prev => 
        prev.map(i => i.id === item.id ? { ...i, name: newName } : i)
      );
      
      toast({
        title: '重命名成功',
        description: `"${item.name}" 已重命名为 "${newName}"`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onDelete: useCallback((itemsToDelete: FileSystemItem[]) => {
      console.log('Deleting items:', itemsToDelete);
      
      const names = itemsToDelete.map(item => item.name);
      
      // Update local state
      setItems(prev => 
        prev.filter(item => !itemsToDelete.some(del => del.id === item.id))
      );
      
      toast({
        title: '删除成功',
        description: `已删除: ${names.join(', ')}`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onDownload: useCallback((itemsToDownload: FileSystemItem[]) => {
      console.log('Downloading items:', itemsToDownload);
      
      const names = itemsToDownload.map(item => item.name);
      
      toast({
        title: '开始下载',
        description: `正在下载: ${names.join(', ')}`,
        status: 'info',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onCreateFolder: useCallback(async (path: string, name: string) => {
      console.log('Creating folder:', name, 'in path:', path);
      
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const newFolder: FileSystemItem = {
        id: `folder_${Date.now()}`,
        name,
        type: 'folder',
        size: 0,
        lastModified: new Date(),
        path: path === '/' ? `/${name}` : `${path}/${name}`,
        parentPath: path,
        filename: name,
        originalFilename: name,
        fileSize: 0,
        mimeType: 'application/x-directory',
        md5: '',
        storageType: 'LOCAL' as any,
        bucket: '',
        objectKey: '',
        status: 'COMPLETE' as any,
        uploaderId: 'user1',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        isPublic: false,
      };
      
      setItems(prev => [...prev, newFolder]);
      
      toast({
        title: '创建成功',
        description: `文件夹 "${name}" 已创建`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    }, [toast]),

    onRefresh: useCallback(() => {
      setIsRefreshing(true);
      loadItems(currentPath).finally(() => {
        setIsRefreshing(false);
      });
    }, [currentPath, loadItems]),
  };

  const breadcrumbs = generateBreadcrumbs(currentPath);

  return (
    <Container maxW="container.xl" py={8}>
      <VStack spacing={6} align="stretch">
        <Box>
          <Heading size="lg" mb={2}>
            文件管理器演示
          </Heading>
          <Box color="gray.600">
            完整的文件管理界面，包含文件浏览、上传、重命名、删除等功能
          </Box>
        </Box>

        <FileExplorer
          currentPath={currentPath}
          items={items}
          breadcrumbs={breadcrumbs}
          uploads={uploads}
          isLoading={isLoading}
          isRefreshing={isRefreshing}
          actions={actions}
          config={{
            allowMultiSelect: true,
            maxFileSize: 50 * 1024 * 1024, // 50MB
            maxConcurrentUploads: 3,
            enableDragAndDrop: true,
            enableContextMenu: true,
          }}
        />
      </VStack>
    </Container>
  );
}
