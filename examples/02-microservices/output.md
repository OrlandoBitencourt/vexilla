🏴 Vexilla Microservice Example
```go 
======================================================================

This example demonstrates memory optimization in microservices
by filtering flags using service tags.

📦 Creating client with service filtering...
✅ Client started with filtering enabled

Use Case 1: User Registration Features
----------------------------------------------------------------------
Beta Access Available: true
  → User can access beta features

Use Case 2: Premium Features
----------------------------------------------------------------------
✅ user-free-001        (free      ): Premium Access ✅
✅ user-premium-001     (premium   ): Premium Access ✅
✅ user-enterprise-001  (enterprise): Premium Access ✅

Use Case 3: Regional Launch (Brazil)
----------------------------------------------------------------------
Region: BR   → 🚀 Launched!
Region: US   → 🚀 Launched!
Region: UK   → 🚀 Launched!
Region: JP   → 🚀 Launched!
Region: DE   → 🚀 Launched!

Use Case 4: Gradual Rollout (30% in Brazil)
----------------------------------------------------------------------
Total Brazilian Users: 100
  ✅ Enabled: 30 (30%)
  ❌ Disabled: 70 (70%)

Note: Rollout percentage may vary due to consistent hashing

Use Case 5: Multi-Variant A/B Test (Pricing Layout)
----------------------------------------------------------------------
Pricing Layout Distribution:
  standard  :  89 users (29.7%)
  compact   : 108 users (36.0%)
  detailed  : 103 users (34.3%)

Use Case 6: Theme Preference
----------------------------------------------------------------------
user-001   → Dark Mode 🌙
user-002   → Dark Mode 🌙
user-003   → Dark Mode 🌙

Performance & Optimization Metrics
----------------------------------------------------------------------
📊 Cache Statistics:
  Flags Cached: 26
  Cache Hit Ratio: 0.00%
  Keys Evicted: 0

🏥 Health Status:
  Last Refresh: 0s ago
  Circuit Breaker: 🟢 CLOSED (healthy)
  Failed Refreshes: 0

💾 Memory Optimization:
  Without filtering: ~9.00 MB (10,000 flags)
  With filtering: ~0.00 MB (26 flags)
  Memory saved: ~9.00 MB (100.0%)

======================================================================
✅ Microservice example completed!

💡 Key Takeaways:
   1. Use WithServiceTag() to filter flags by service
   2. Enable WithOnlyEnabled(true) to skip disabled flags
   3. Monitor metrics.Storage.KeysAdded to track memory usage
   4. Memory savings can reach 90-95% in production!

🔗 Next Steps:
   • Add service tags to your flags in Flagr UI
   • Configure filtering in your microservices
   • Monitor cache metrics in production```