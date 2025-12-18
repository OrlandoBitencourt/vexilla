🏴 Vexilla Basic Example
======================================================================

📦 Creating Vexilla client...
🚀 Starting client and loading flags...
✅ Client started successfully!

Example 1: Boolean Flag Evaluation
----------------------------------------------------------------------
Flag: new_feature
User: user-123
Result: true

Example 2: Boolean Flag with Constraints
----------------------------------------------------------------------
Flag: premium_features
User: user-456 (tier=premium)
Result: true

Example 3: String Flag (UI Theme)
----------------------------------------------------------------------
Flag: ui_theme
User: user-789
Theme: light

Example 4: Integer Flag (Max Items)
----------------------------------------------------------------------
Flag: max_items
User: user-999
Max items: 100

Example 5: Detailed Evaluation
----------------------------------------------------------------------
Flag Key: dark_mode
Variant: enabled
Is Enabled: true
Reason: matched segment 2

Example 6: Multiple User Contexts
----------------------------------------------------------------------
user-001     (tier=free      ): true
user-002     (tier=premium   ): true
user-003     (tier=enterprise): true

Example 7: A/B Test Distribution
----------------------------------------------------------------------
Button Color A/B Test Results:
  blue: 40% (40 users)
  red: 60% (60 users)

Example 8: Performance Metrics
----------------------------------------------------------------------
Cache Performance:
  Keys Cached: 30
  Keys Evicted: 0
  Hit Ratio: 0.00%
  Last Refresh: 0s ago
  Circuit Open: false
  Consecutive Fails: 0

======================================================================
✅ Example completed successfully!

💡 Try these commands:
   • Modify flags in Flagr UI: http://localhost:18000
   • Run this example again to see changes
   • Check internal/cache metrics for performance data