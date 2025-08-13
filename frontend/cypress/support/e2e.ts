// Import commands.ts using ES2015 syntax:
import './commands'

// Alternatively you can use CommonJS syntax:
// require('./commands')

import '@cypress/code-coverage/support'

// Configure global test settings
beforeEach(() => {
  // Preserve authentication across tests
  cy.window().then((win) => {
    win.sessionStorage.clear()
  })
  
  // Set up API intercepts for consistent testing
  cy.intercept('GET', '/api/v1/health', { statusCode: 200, body: { status: 'ok' } })
  
  // Mock WebDAV endpoints for testing
  cy.intercept('GET', '/api/v1/*/files**', { 
    statusCode: 200, 
    body: { 
      items: [], 
      total: 0,
      hasMore: false 
    } 
  }).as('listFiles')
  
  cy.intercept('POST', '/api/v1/*/file**', { 
    statusCode: 200, 
    body: { 
      success: true,
      message: 'File uploaded successfully' 
    } 
  }).as('uploadFile')
  
  cy.intercept('DELETE', '/api/v1/**', { 
    statusCode: 200, 
    body: { 
      success: true,
      message: 'File deleted successfully' 
    } 
  }).as('deleteFile')
  
  cy.intercept('PATCH', '/api/v1/*/rename', { 
    statusCode: 200, 
    body: { 
      success: true,
      message: 'File renamed successfully' 
    } 
  }).as('renameFile')
})

// Global error handling
Cypress.on('uncaught:exception', (err, runnable) => {
  // Don't fail tests on React hydration errors or other non-critical errors
  if (err.message.includes('Hydration') || 
      err.message.includes('ResizeObserver') ||
      err.message.includes('Non-Error promise rejection captured')) {
    return false
  }
  return true
})

// Performance monitoring
let performanceEntries: PerformanceEntry[] = []

beforeEach(() => {
  performanceEntries = []
  
  cy.window().then((win) => {
    // Clear previous performance entries
    if (win.performance && win.performance.clearMarks) {
      win.performance.clearMarks()
      win.performance.clearMeasures()
    }
  })
})

afterEach(() => {
  cy.window().then((win) => {
    if (win.performance && win.performance.getEntriesByType) {
      performanceEntries = win.performance.getEntriesByType('navigation')
      
      // Log performance metrics
      if (performanceEntries.length > 0) {
        const navigation = performanceEntries[0] as PerformanceNavigationTiming
        cy.task('log', `Page load time: ${navigation.loadEventEnd - navigation.navigationStart}ms`)
        cy.task('log', `DOM content loaded: ${navigation.domContentLoadedEventEnd - navigation.navigationStart}ms`)
      }
    }
  })
})
