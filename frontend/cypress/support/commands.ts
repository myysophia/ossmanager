/// <reference types="cypress" />

import '@cypress/code-coverage/support'

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Login with credentials and get auth token
       */
      login(username?: string, password?: string): Chainable<void>
      
      /**
       * Login with admin privileges
       */
      loginAsAdmin(): Chainable<void>
      
      /**
       * Setup test bucket and permissions
       */
      setupTestBucket(bucketName: string): Chainable<void>
      
      /**
       * Upload a test file via drag and drop
       */
      uploadFile(fileName: string, fileContent: string, mimeType?: string): Chainable<void>
      
      /**
       * Upload large file for testing
       */
      uploadLargeFile(fileName: string, sizeMB: number): Chainable<void>
      
      /**
       * Navigate to file explorer
       */
      navigateToFileExplorer(bucket?: string, path?: string): Chainable<void>
      
      /**
       * Wait for file operation to complete
       */
      waitForFileOperation(): Chainable<void>
      
      /**
       * Verify file exists in the list
       */
      verifyFileExists(fileName: string): Chainable<void>
      
      /**
       * Simulate network failure
       */
      simulateNetworkFailure(): Chainable<void>
      
      /**
       * Restore network
       */
      restoreNetwork(): Chainable<void>
      
      /**
       * Create test data files
       */
      createTestFile(name: string, size: number): Chainable<File>
    }
  }
}

// Authentication commands
Cypress.Commands.add('login', (username = 'testuser', password = 'testpass123') => {
  cy.session([username, password], () => {
    cy.request({
      method: 'POST',
      url: '/api/auth/login',
      body: {
        username,
        password
      },
      failOnStatusCode: false
    }).then((resp) => {
      if (resp.status === 200) {
        window.localStorage.setItem('auth-token', resp.body.token)
        window.localStorage.setItem('user-info', JSON.stringify(resp.body.user))
      }
    })
  })
})

Cypress.Commands.add('loginAsAdmin', () => {
  cy.login('admin', 'admin123')
})

// Test setup commands
Cypress.Commands.add('setupTestBucket', (bucketName: string) => {
  cy.request({
    method: 'POST',
    url: '/api/v1/buckets',
    body: {
      name: bucketName,
      region: 'test-region',
      provider: 'mock'
    },
    headers: {
      'Authorization': `Bearer ${window.localStorage.getItem('auth-token')}`
    }
  })
})

// File operation commands
Cypress.Commands.add('uploadFile', (fileName: string, fileContent: string, mimeType = 'text/plain') => {
  cy.get('[data-testid="file-drop-zone"]').should('be.visible')
  
  // Create a test file
  const file = new File([fileContent], fileName, { type: mimeType })
  
  // Trigger drop event
  cy.get('[data-testid="file-drop-zone"]').selectFile({
    contents: Cypress.Buffer.from(fileContent),
    fileName,
    mimeType
  }, { action: 'drag-drop' })
  
  cy.waitForFileOperation()
})

Cypress.Commands.add('uploadLargeFile', (fileName: string, sizeMB: number) => {
  // Create large file content (pattern-based for memory efficiency)
  const chunkSize = 1024 * 1024 // 1MB chunks
  const totalSize = sizeMB * chunkSize
  
  cy.createTestFile(fileName, totalSize).then((file) => {
    cy.get('[data-testid="file-drop-zone"]').selectFile(file, { action: 'drag-drop' })
    cy.waitForFileOperation()
  })
})

Cypress.Commands.add('navigateToFileExplorer', (bucket = 'test-bucket', path = '/') => {
  cy.visit(`/files/${bucket}?path=${encodeURIComponent(path)}`)
  cy.get('[data-testid="file-explorer"]').should('be.visible')
})

Cypress.Commands.add('waitForFileOperation', () => {
  // Wait for upload progress to appear and disappear
  cy.get('[data-testid="upload-progress"]', { timeout: 30000 }).should('be.visible')
  cy.get('[data-testid="upload-progress"]', { timeout: 60000 }).should('not.exist')
  
  // Wait for file list to refresh
  cy.get('[data-testid="file-list"]').should('be.visible')
})

Cypress.Commands.add('verifyFileExists', (fileName: string) => {
  cy.get('[data-testid="file-list"]')
    .contains('[data-testid="file-item"]', fileName)
    .should('be.visible')
})

// Network simulation commands
Cypress.Commands.add('simulateNetworkFailure', () => {
  cy.intercept('**', { statusCode: 500, body: { error: 'Network failure simulation' } }).as('networkFailure')
})

Cypress.Commands.add('restoreNetwork', () => {
  cy.intercept('**').as('networkRestored')
})

// Utility commands
Cypress.Commands.add('createTestFile', (name: string, size: number) => {
  return cy.then(() => {
    // Create file with pattern data to be memory efficient
    const chunks: string[] = []
    const pattern = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    const chunkSize = 8192 // 8KB chunks
    
    for (let i = 0; i < size; i += chunkSize) {
      const remainingSize = Math.min(chunkSize, size - i)
      let chunk = ''
      for (let j = 0; j < remainingSize; j++) {
        chunk += pattern[(i + j) % pattern.length]
      }
      chunks.push(chunk)
    }
    
    const content = chunks.join('')
    return new File([content], name, { type: 'application/octet-stream' })
  })
})

export {}
