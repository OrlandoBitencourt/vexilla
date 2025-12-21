# Deterministic Rollout Example

This example demonstrates how to achieve **100% local evaluation** for rollouts using pre-computed buckets instead of random percentage-based evaluations.

## Problem with Traditional Rollouts

Traditional percentage-based rollouts have several drawbacks:
- ❌ **Non-deterministic**: Random evaluation means results can change between calls
- ❌ **HTTP-dependent**: Requires Flagr API calls for consistent hashing
- ❌ **Hard to debug**: Difficult to reproduce user-specific behavior
- ❌ **Network latency**: Adds 50-200ms per evaluation

## Solution: Pre-computed Buckets

Instead of relying on Flagr's random rollout percentages, we pre-compute a **deterministic bucket** from user identifiers and send it as a simple numeric attribute.

## How It Works

```
User Identifier (CPF, UserID, etc.)
    │
    ▼
Application (pre-processing)
    │
    ├─> Extract/Hash identifier
    │   └─> Generate bucket: 0-99
    │
    ▼
Vexilla.Evaluate(ctx, flag, {user_bucket: 42})
    │
    ▼
Local Evaluation (no HTTP)
    │
    ├─> Check: user_bucket >= 0 AND user_bucket <= 69?
    │   └─> Match: Enabled (70% rollout)
    │
    └─> Else: Disabled
```

## Running the Example

```bash
cd examples/03-deterministic-rollout
go run main.go
```

## Flagr Configuration

### Example: 70% Rollout

**Segment 1 (70% rollout):**
- Constraint 1: `user_bucket` >= `0`
- Constraint 2: `user_bucket` <= `69`
- Variant: Enabled

**Segment 2 (Default):**
- No constraints
- Variant: Disabled

## Performance Benefits

| Approach | Latency | HTTP Calls | Deterministic |
|----------|---------|------------|---------------|
| Flagr % rollout | 50-200ms | 1 per eval | ❌ Random |
| Deterministic bucket | <1ms | 0 | ✅ Stable |

**Speedup**: ~50-200x faster, zero network overhead

## Use Cases

- ✅ **Critical features**: Where latency matters
- ✅ **Edge computing**: Limited connectivity environments
- ✅ **Mobile apps**: Reduce battery drain from network calls
- ✅ **Compliance**: Reproducible audit trails
- ✅ **A/B testing**: Stable user assignments without sticky sessions
