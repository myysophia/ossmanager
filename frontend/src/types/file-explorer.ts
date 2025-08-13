/**
 * File Explorer types and interfaces
 */
import { OSSFile } from './file';

/**
 * File system item (can be file or folder)
 */
export interface FileSystemItem extends Omit<OSSFile, 'id'> {
  id: string;
  name: string;
  type: 'file' | 'folder';
  size: number;
  lastModified: Date;
  path: string;
  parentPath?: string;
  children?: FileSystemItem[];
  isLoading?: boolean;
}

/**
 * Breadcrumb item for navigation
 */
export interface BreadcrumbItem {
  name: string;
  path: string;
  isCurrentPage?: boolean;
}

/**
 * Upload progress information
 */
export interface UploadProgress {
  id: string;
  file: File;
  progress: number;
  status: 'uploading' | 'completed' | 'failed' | 'cancelled';
  error?: string;
}

/**
 * Context menu item
 */
export interface ContextMenuItem {
  id: string;
  label: string;
  icon?: React.ComponentType;
  action: () => void;
  disabled?: boolean;
  isDanger?: boolean;
}

/**
 * File explorer configuration
 */
export interface FileExplorerConfig {
  allowMultiSelect?: boolean;
  showHiddenFiles?: boolean;
  allowedFileTypes?: string[];
  maxFileSize?: number;
  maxConcurrentUploads?: number;
  enableDragAndDrop?: boolean;
  enableContextMenu?: boolean;
}

/**
 * File explorer actions
 */
export interface FileExplorerActions {
  onNavigate: (path: string) => void;
  onFileSelect: (file: FileSystemItem) => void;
  onFileDoubleClick: (file: FileSystemItem) => void;
  onFolderDoubleClick: (folder: FileSystemItem) => void;
  onUpload: (files: File[], path: string) => void;
  onRename: (item: FileSystemItem, newName: string) => void;
  onDelete: (items: FileSystemItem[]) => void;
  onDownload: (items: FileSystemItem[]) => void;
  onCreateFolder: (path: string, name: string) => void;
  onRefresh: () => void;
}

/**
 * Sort configuration
 */
export interface SortConfig {
  key: keyof FileSystemItem;
  direction: 'asc' | 'desc';
}

/**
 * Table column configuration
 */
export interface TableColumn {
  id: string;
  header: string;
  accessor: keyof FileSystemItem;
  width?: string;
  sortable?: boolean;
  hidden?: boolean;
}
