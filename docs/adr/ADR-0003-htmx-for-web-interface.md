# ADR-0003: Use HTMX for Web Interface

- Status: accepted
- Date: 2026-01-08
- Deciders: Sergey Grachev

## Context

The KBase Image Catalog project requires a web interface for browsing image catalogs, performing searches, and managing the cataloging process. The frontend technology choice impacts development velocity, user experience, maintainability, and performance.

Key considerations:
- **Development Velocity**: Fast iteration and easy maintenance
- **User Experience**: Dynamic updates without full page reloads
- **Performance**: Fast loading times and minimal bandwidth usage
- **Learning Curve**: Team familiarity and onboarding time
- **Ecosystem**: Available libraries, tools, and community support
- **SEO Requirements**: Search engine optimization considerations
- **Mobile Responsiveness**: Cross-device compatibility
- **Accessibility**: WCAG compliance and screen reader support

## Decision

We will use **HTMX** as the primary frontend technology for the KBase Image Catalog web interface, combined with standard HTML, CSS, and minimal JavaScript.

## Consequences

### Positive Consequences

1. **Simplified Development**: Minimal JavaScript knowledge required, leverages existing HTML/CSS skills
2. **Progressive Enhancement**: Works with basic HTML, enhances with JavaScript when available
3. **Reduced Complexity**: No complex JavaScript framework to learn and maintain
4. **Fast Performance**: Minimal JavaScript payload, fast page loads
3. **SEO Friendly**: Server-rendered HTML content is easily crawlable by search engines
4. **Accessibility**: Standard HTML elements are inherently accessible
5. **Easy Testing**: HTML templates are easy to unit test and debug
6. **Wide Browser Support**: Works in all modern browsers without polyfills
7. **Small Learning Curve**: Developers can be productive quickly
8. **Reduced Build Complexity**: No complex build tools or bundling required
9. **Direct Server Communication**: HTMX makes direct AJAX calls to the backend
10. **Rich Interactions**: Supports advanced features like WebSockets, Server-Sent Events, and transitions

### Negative Consequences

1. **Limited Rich UI Components**: Fewer pre-built components compared to major frameworks
2. **JavaScript Dependency**: Still requires some JavaScript for advanced interactions
3. **State Management**: No built-in state management solution
4. **Tooling Maturity**: Less mature tooling ecosystem compared to React/Vue/Angular
5. **Debugging Complexity**: Errors can be harder to trace between server and client
6. **Performance Trade-offs**: Client-side rendering can be more performant for complex SPAs
7. **Team Expertise**: Limited pool of developers experienced with HTMX
8. **Documentation**: Less comprehensive documentation compared to major frameworks

## Alternatives Considered

### 1. React
**Rejected** because:
- High complexity and learning curve
- Requires build tools and bundling
- Larger bundle size and slower initial load
- More moving parts (state management, routing, etc.)
- Overkill for the catalog browsing requirements

### 2. Vue.js
**Rejected** because:
- Still requires significant JavaScript knowledge
- Build tooling complexity
- Larger than HTMX for the use case
- More state management complexity than needed

### 3. Angular
**Rejected** because:
- Very high complexity and learning curve
- Large framework with many dependencies
- Steep learning curve for team
- Over-engineered for the catalog requirements

### 4. Svelte
**Rejected** because:
- Requires compilation step
- Smaller ecosystem than major frameworks
- Team familiarity limitations
- Build tooling complexity

### 5. Plain HTML + Vanilla JavaScript
**Rejected** because:
- Would require significant custom code for dynamic updates
- More development effort for similar functionality
- Less maintainable than HTMX approach
- Would reinvent wheel for common interactions

### 6. Server-Side Templates Only
**Rejected** because:
- Poor user experience with full page reloads
- No real-time updates
- Less responsive interface
- Would require more server roundtrips

## Related ADRs

- [ADR-0002: Use Go for Backend Development](./ADR-0002-use-go-for-backend.md) - Backend technology choice
- [ADR-0004: Integrate LLM Models for Image Analysis](./ADR-0004-llm-integration.md) - AI integration that affects UI
- [ADR-0005: Adopt Standard Go Project Structure](./ADR-0005-project-structure.md) - Template organization
- [ADR-0007: Focus on Performance Optimization](./ADR-0007-performance-optimization.md) - Performance considerations for UI

## Implementation Details

### HTMX Features Utilized

1. **HTTP Verbs**: Use of GET, POST, PUT, DELETE for different operations
2. **Triggers**: Event-driven updates (click, form submission, etc.)
3. **Swaps**: Different swap strategies for different content types
4. **Headers**: Custom headers for authentication and content negotiation
5. **Events**: Custom events for complex interactions
6. **WebSockets**: Real-time updates for catalog changes
7. **Server-Sent Events**: Live notifications and updates

### Template Organization

```
web/
├── templates/
│   ├── index.html              # Main catalog listing
│   ├── catalog-detail.html     # Individual catalog view
│   ├── search-result.html      # Search results
│   ├── catalog-list-fragment.html # Reusable list component
│   ├── catalog-images-fragment.html # Image grid component
│   ├── search-result-fragment.html # Search results component
│   └── catalog-navigation-template.html # Navigation component
└── static/
    ├── styles.css              # Custom CSS styles
    ├── htmx.min.js             # HTMX library
    ├── viewer.min.js           # Image viewer library
    └── viewer.min.css          # Image viewer styles
```

### Key Interactions

1. **Catalog Browsing**: HTMX-enabled pagination and filtering
2. **Search**: Real-time search with debounced requests
3. **Image Preview**: Modal popups for image details
4. **Auto-refresh**: Server-sent events for catalog updates
5. **Infinite Scroll**: Load more images as user scrolls

### Performance Optimizations

1. **Lazy Loading**: Images load only when visible
2. **Debounced Search**: Prevent excessive API calls
3. **Caching**: Browser caching for static assets
4. **Compression**: Gzip compression for HTML responses
5. **Minification**: Minified CSS and JavaScript

## Future Considerations

- Monitor HTMX evolution and adoption trends
- Evaluate integration with emerging web technologies
- Consider accessibility enhancements as project matures
- Plan for potential offline functionality
- Evaluate progressive web app capabilities

## Testing Strategy

1. **Unit Tests**: Test individual template components
2. **Integration Tests**: Test HTMX interactions with backend
3. **E2E Tests**: Test complete user workflows
4. **Performance Tests**: Measure load times and interactivity
5. **Accessibility Tests**: Ensure WCAG compliance
6. **Browser Compatibility**: Test across target browsers