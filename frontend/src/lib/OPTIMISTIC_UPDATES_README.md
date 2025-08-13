# Optimistic Updates System

This document explains the optimistic updates implementation with SWR for the file management system.

## Overview

The optimistic updates system provides instant UI feedback by immediately updating the client state when users perform actions, then rolling back if the server operation fails. This creates a responsive user experience while maintaining data consistency.

## Key Features

- **Immediate UI Updates**: Actions appear instantly in the UI
- **Automatic Rollback**: Failed operations are automatically reverted
- **Centralized Toast Notifications**: Consistent success/error messaging
- **SWR Integration**: Efficient caching and revalidation
- **Error Handling**: Comprehensive error recovery

## Architecture

### Components

1. **Toast System** (`src/lib/toast.ts`)
   - Centralized notification management
   - Consistent toast styling and behavior
   - Utility functions for common scenarios

2. **Optimistic Update Manager** (`src/lib/optimistic-updates.ts`)
   - Manages optimistic state changes
   - Handles rollback operations
   - Tracks pending actions

3. **SWR Hooks** (`src/lib/hooks/useFilesWithSWR.ts`)
   - Data fetching with caching
   - Optimistic cache updates
   - Automatic revalidation

4. **UI Components**
   - `FileUploadZone`: Drag & drop upload with progress
   - `EnhancedFileListPage`: Complete file management interface

## Usage Examples

### File Upload with Optimistic Updates

```typescript
import { useFilesWithSWR } from '../lib/hooks/useFilesWithSWR';

const { uploadFile } = useFilesWithSWR();

const handleUpload = async (file: File) => {
  try {
    // Immediately shows file in UI, then updates with real data
    await uploadFile(file);
  } catch (error) {
    // Automatically rolls back and shows error toast
    console.error('Upload failed:', error);
  }
};
```

### File Deletion with Optimistic Updates

```typescript
const { deleteFile } = useFilesWithSWR();

const handleDelete = async (fileId: number) => {
  try {
    // File immediately disappears from UI
    await deleteFile(fileId);
  } catch (error) {
    // File reappears if deletion failed
    console.error('Delete failed:', error);
  }
};
```

### Batch Operations

```typescript
const { deleteFiles } = useFilesWithSWR();

const handleBatchDelete = async (fileIds: number[]) => {
  try {
    // All selected files immediately disappear
    await deleteFiles(fileIds);
  } catch (error) {
    // Files reappear if operation failed
    console.error('Batch delete failed:', error);
  }
};
```

## Flow Diagram

```
User Action
    ↓
Optimistic Update (Immediate UI Change)
    ↓
API Call
    ↓
Success? → Yes → Complete Update + Success Toast
    ↓
    No → Rollback + Error Toast + Revalidate
```

## Implementation Details

### 1. Starting an Optimistic Update

```typescript
const updateKey = optimisticUpdateManager.startUpdate(
  'delete',                    // Action type
  fileId,                     // Target identifier
  () => {                     // Optimistic update function
    setFiles(prev => prev.filter(f => f.id !== fileId));
  },
  () => {                     // Rollback function
    setFiles(prev => [...prev, deletedFile]);
  }
);
```

### 2. Completing/Failing Updates

```typescript
try {
  await FileAPI.deleteFile(fileId);
  optimisticUpdateManager.completeUpdate(updateKey);
} catch (error) {
  optimisticUpdateManager.rollbackUpdate(updateKey, error.message);
}
```

### 3. SWR Cache Integration

```typescript
// Update SWR cache optimistically
mutate(key, {
  ...currentData,
  items: updatedItems,
  total: newTotal,
}, false); // Don't revalidate immediately

// Revalidate on error
if (error) {
  revalidate(); // Fetch fresh data from server
}
```

## Toast Notifications

### Success Notifications
- ✅ Upload successful
- ✅ File deleted
- ✅ Download started

### Error Notifications
- ❌ Upload failed with reason
- ❌ Delete operation failed
- ❌ Network errors
- ❌ Permission errors

### Progress Notifications
- 📤 Upload progress (0% - 100%)
- 🔄 Processing indicators

## SWR Configuration

```typescript
const { data, error, mutate } = useSWR(
  key,
  fetcher,
  {
    revalidateOnFocus: false,    // Don't refetch on window focus
    revalidateOnReconnect: true, // Refetch on network reconnect
    errorRetryCount: 3,          // Retry failed requests 3 times
    errorRetryInterval: 1000,    // Wait 1s between retries
    dedupingInterval: 5000,      // Avoid duplicate requests within 5s
  }
);
```

## Error Recovery

### Automatic Recovery
- Failed operations automatically rollback
- SWR revalidates data on error
- Users see appropriate error messages

### Manual Recovery
- Refresh button forces data revalidation
- Retry mechanisms for failed uploads
- Clear pending operations

## Best Practices

### 1. Always Provide Rollback
```typescript
// ✅ Good: Always provide rollback function
const updateKey = optimisticUpdateManager.startUpdate(
  'action',
  target,
  optimisticUpdate,
  rollbackFunction // Always provide this
);

// ❌ Bad: No rollback function
const updateKey = optimisticUpdateManager.startUpdate(
  'action',
  target,
  optimisticUpdate
);
```

### 2. Handle All Error Cases
```typescript
try {
  await apiOperation();
  optimisticUpdateManager.completeUpdate(key);
} catch (error) {
  // ✅ Always handle errors
  optimisticUpdateManager.rollbackUpdate(key, error.message);
  
  // ✅ Revalidate to ensure consistency
  revalidate();
}
```

### 3. Provide User Feedback
```typescript
// ✅ Show loading states
<Button 
  isLoading={optimisticUpdateManager.isPending('delete', fileId)}
  onClick={handleDelete}
>
  Delete
</Button>

// ✅ Show pending operations
{pendingActions.length > 0 && (
  <Alert status="info">
    Processing {pendingActions.length} operations...
  </Alert>
)}
```

## Troubleshooting

### Common Issues

1. **Stale Data After Error**
   - Ensure `revalidate()` is called after errors
   - Check SWR key consistency

2. **Missing Rollback**
   - Always provide rollback functions
   - Test error scenarios

3. **Duplicate Requests**
   - Use SWR deduplication
   - Check for proper key generation

### Debugging

```typescript
// Check pending actions
console.log(optimisticUpdateManager.getPendingActions());

// Monitor SWR cache
console.log(cache.get(key));

// Track rollbacks
optimisticUpdateManager.on('rollback', (key, error) => {
  console.log('Rollback:', key, error);
});
```

## Performance Considerations

- **SWR Caching**: Reduces redundant API calls
- **Optimistic Updates**: Improves perceived performance
- **Background Revalidation**: Keeps data fresh
- **Request Deduplication**: Prevents duplicate requests

## Testing

### Unit Tests
```typescript
describe('Optimistic Updates', () => {
  it('should rollback on failure', async () => {
    const initialState = [file1, file2];
    const { result } = renderHook(() => useFilesWithSWR());
    
    // Mock API failure
    mockAPI.delete.mockRejectedValue(new Error('Failed'));
    
    await act(async () => {
      await result.current.deleteFile(file1.id);
    });
    
    // Should rollback to initial state
    expect(result.current.files).toEqual(initialState);
  });
});
```

### Integration Tests
- Test complete user workflows
- Verify error recovery
- Check toast notifications

## Migration Guide

### From Traditional Hooks

```typescript
// Before: Traditional hook
const { deleteFile } = useFiles();

// After: SWR with optimistic updates  
const { deleteFile } = useFilesWithSWR();
// Usage remains the same, but now with optimistic updates
```

### Adding to Existing Components

1. Replace traditional hooks with SWR versions
2. Remove manual loading states (handled by SWR)
3. Remove manual error handling (centralized in hooks)
4. Add pending action indicators

## Future Enhancements

- [ ] Offline support with queue
- [ ] Conflict resolution
- [ ] Advanced retry strategies
- [ ] Real-time updates with WebSockets
- [ ] Optimistic updates for more operations
