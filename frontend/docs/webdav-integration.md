# WebDAV Token Picker and URL Helper Integration

This document describes the integration of token picker and WebDAV URL helper components into the WebDAV browser functionality.

## Overview

The integration provides:
- **Reusable token management** from the `/main/webdav` page
- **Token selection/creation** interface for browser access
- **Quick WebDAV URL helper** for external client mounting
- **Seamless UX** that distinguishes between browser (JWT) and external client (token) access

## Components

### 1. TokenPicker Component

**Location**: `src/components/webdav/TokenPicker.tsx`

**Purpose**: Reusable component for selecting and managing WebDAV tokens

**Features**:
- Load tokens for a specific bucket
- Create new tokens with expiration settings
- Auto-select valid tokens
- Token status display (valid/expired)
- Copy WebDAV URL functionality
- Responsive design with size options

**Props**:
```typescript
interface TokenPickerProps {
  selectedBucket: string;
  onTokenSelect?: (token: WebDAVToken | null) => void;
  showWebDAVUrl?: boolean;
  showCreateOption?: boolean;
  size?: 'sm' | 'md' | 'lg';
}
```

**Usage**:
```tsx
<TokenPicker
  selectedBucket={bucketName}
  onTokenSelect={setSelectedToken}
  showWebDAVUrl={false}
  size="md"
/>
```

### 2. WebDAVUrlHelper Component

**Location**: `src/components/webdav/WebDAVUrlHelper.tsx`

**Purpose**: Helper for external WebDAV client access

**Features**:
- Display WebDAV URL with copy functionality
- Token status validation
- Quick action buttons (copy URL, copy token, open client)
- Usage instructions for external clients
- Token expiration warnings

**Props**:
```typescript
interface WebDAVUrlHelperProps {
  selectedBucket: string;
  selectedToken?: WebDAVToken | null;
  size?: 'sm' | 'md' | 'lg';
  showInstructions?: boolean;
}
```

**Usage**:
```tsx
<WebDAVUrlHelper
  selectedBucket={bucketName}
  selectedToken={selectedToken}
  showInstructions={true}
/>
```

### 3. useWebDAVTokens Hook

**Location**: `src/hooks/useWebDAVTokens.ts`

**Purpose**: Custom hook for WebDAV token state management

**Features**:
- Token CRUD operations
- Auto-refresh functionality
- Bucket-specific token filtering
- Token validation helpers
- State management with error handling

**Usage**:
```tsx
const {
  tokens,
  selectedToken,
  loading,
  selectToken,
  createToken,
  deleteToken,
  isTokenExpired,
  getWebDAVUrl
} = useWebDAVTokens({ 
  bucket: selectedBucket,
  autoRefresh: true 
});
```

## Integration Points

### 1. WebDAV Browser Page

**Location**: `src/app/main/webdav/browser/page.tsx`

**Integration**:
- Added `TokenPicker` and `WebDAVUrlHelper` components
- Side-by-side layout on desktop, stacked on mobile
- Token state management with `useState`
- Real-time status updates

**Layout**:
```jsx
{selectedBucket && (
  <Grid templateColumns={{ base: '1fr', lg: '1fr 1fr' }} gap={6}>
    <GridItem>
      <TokenPicker
        selectedBucket={selectedBucket}
        onTokenSelect={setSelectedToken}
        showWebDAVUrl={false}
      />
    </GridItem>
    <GridItem>
      <WebDAVUrlHelper
        selectedBucket={selectedBucket}
        selectedToken={selectedToken}
        showInstructions={true}
      />
    </GridItem>
  </Grid>
)}
```

### 2. Browser Status Display

**Enhanced status indicators**:
- ✅ Full configuration (bucket + token available)
- ⚠️ Browser-only access (bucket selected, no token)
- Clear distinction between browser and external client access

## User Experience Flow

### 1. Browser Access (Proxy API)
1. User selects bucket from dropdown
2. Browser automatically uses JWT authentication
3. No token required for browser-based file operations
4. WebDAV proxy API handles authentication seamlessly

### 2. External Client Access
1. User selects bucket from dropdown
2. TokenPicker component loads available tokens
3. User selects existing token or creates new one
4. WebDAVUrlHelper displays connection information
5. User copies URL and token to external client
6. Client connects using username/token authentication

## Technical Details

### Authentication Flow
- **Browser**: JWT token from localStorage → WebDAV proxy API
- **External**: Username + WebDAV access token → Direct WebDAV endpoint

### Token Management
- Tokens are scoped to specific buckets
- Automatic expiration handling
- Real-time validation and status updates
- Secure token creation with configurable expiration

### URL Generation
```javascript
const webdavUrl = `${window.location.origin}/webdav/${bucketName}`;
```

### Error Handling
- Network failures with retry options
- Token validation with clear messaging
- Graceful degradation for missing permissions

## Configuration Options

### TokenPicker Configuration
- `showWebDAVUrl`: Show/hide URL in picker (default: true)
- `showCreateOption`: Enable/disable token creation (default: true)
- `size`: Component size variation (sm/md/lg)

### WebDAVUrlHelper Configuration
- `showInstructions`: Display usage instructions (default: true)
- Token status validation with expiration checks
- Quick action buttons for common operations

## Benefits

### 1. Unified Token Management
- Reuses existing token infrastructure from `/main/webdav`
- Consistent token creation and management
- Centralized token validation logic

### 2. Improved User Experience
- Clear distinction between browser and external access
- Quick access to WebDAV URLs with copy functionality
- Real-time token status updates
- Intuitive interface for both novice and expert users

### 3. Development Benefits
- Reusable components for future WebDAV features
- Consistent API usage patterns
- Type-safe interfaces with TypeScript
- Comprehensive error handling

## Future Enhancements

1. **Token Templates**: Pre-configured token settings for common use cases
2. **Client Detection**: Automatic client configuration based on user agent
3. **Bulk Operations**: Multi-bucket token management
4. **Token Sharing**: Secure token sharing with expiration controls
5. **Usage Analytics**: Token usage tracking and statistics
