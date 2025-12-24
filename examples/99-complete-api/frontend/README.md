# Vexilla Demo - Frontend

Next.js + TypeScript frontend for visualizing feature flags in action.

## Quick Start

```bash
# Install dependencies
npm install

# Run development server
npm run dev
```

Frontend runs on `http://localhost:3000`

## Features

- ✅ User simulator for testing different contexts
- ✅ Real-time flag status visualization
- ✅ Rollout indicator with visual explanation
- ✅ Checkout demo (V1 vs V2)
- ✅ Admin panel for cache invalidation

## Environment Variables

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Components

### UserSimulator

Allows switching between mock users or creating custom user contexts.

```tsx
<UserSimulator
  initialContext={context}
  onContextChange={handleContextChange}
/>
```

### FlagStatus

Displays current state of a feature flag.

```tsx
<FlagStatus
  name="api.checkout.v2"
  value={true}
  type="boolean"
/>
```

### RolloutIndicator

Visual explanation of deterministic rollout algorithm.

```tsx
<RolloutIndicator
  cpf="12345678909"
  bucket={42}
  rollout={30}
  enabled={true}
/>
```

### CheckoutDemo

Shows checkout V1 or V2 based on flag evaluation.

```tsx
<CheckoutDemo
  checkoutData={response}
  loading={false}
/>
```

### AdminActions

Admin operations (only visible when role=admin).

```tsx
<AdminActions
  onInvalidate={refreshFlags}
/>
```

## Development

```bash
# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint code
npm run lint
```

## Styling

Uses Tailwind CSS for styling. Configuration in `tailwind.config.ts`.

## API Integration

API client in `src/services/api.ts` handles all backend communication with proper headers:

- `X-User-ID`
- `X-CPF`
- `X-User-Role`
- `X-Country`
