# Deterministic Rollout (Bucket-based Evaluation)

This example demonstrates how to implement **deterministic rollouts** in Vexilla using pre-processed numeric attributes, removing the need for random percentage-based rollouts and enabling **100% local evaluation** with no HTTP dependency.

---

## 🎯 Motivation

Traditional percentage-based rollouts have several drawbacks:

* **Non-deterministic** evaluation (random)
* Results may change between executions
* Hard to reproduce bugs
* Requires HTTP calls to evaluate decisions

This example replaces random rollouts with a **deterministic bucket**, derived from a user identifier (e.g. CPF).

---

## 🧠 Concept

Instead of:

* 70% of users → Audience A (random)
* 30% → Default

We use:

* Bucket `00–69` → Audience A
* Bucket `70–99` → Default

The bucket is calculated **inside the application**, and Flagr/Vexilla only evaluates a **numeric attribute**.

---

## 🏗️ Architecture

```
CPF (string)
   ↓
Application (pre-processing)
   ↓
cpf_bucket = 0..99
   ↓
Vexilla (local evaluation)
```

📌 Flagr does **not** need to understand CPF, regex, or complex strings.

---

## 1️⃣ CPF Bucket Pre-processing

We extract the 6th and 7th digits of the CPF to generate a bucket between `00` and `99`.

```go
func CPFBucket(cpf string) int {
	clean := strings.NewReplacer(".", "", "-", "").Replace(cpf)

	if len(clean) < 7 {
		return -1
	}

	bucket, err := strconv.Atoi(clean[5:7])
	if err != nil {
		return -1
	}

	return bucket
}
```

---

## 2️⃣ Attributes passed to Vexilla

Only a numeric attribute is sent.

```go
attrs := vexilla.Attributes{
	"cpf_bucket": CPFBucket("123.456.789-09"),
}
```

---

## 3️⃣ Segment configuration in Flagr

### Segment: `audience_a`

Rules:

| Field      | Operator | Value |
| ---------- | -------- | ----- |
| cpf_bucket | >=       | 0     |
| cpf_bucket | <=       | 69    |

📌 No regex is used.
📌 Only standard numeric operators supported by Flagr.

---

## 4️⃣ Flag configuration

* Segment: `audience_a`
* Rollout: **100%**
* Default: `disabled`

⚠️ The percentage is no longer random; it acts as a logical match.

---

## 5️⃣ Local evaluation with Vexilla

After Vexilla synchronizes flags:

```go
if vexilla.IsEnabled("new-experience", attrs) {
	// Audience A
} else {
	// Default
}
```

✅ Deterministic evaluation
✅ No HTTP calls
✅ Fully cacheable

---

## 🧪 Practical examples

| CPF            | Bucket | Result     |
| -------------- | ------ | ---------- |
| 123.450.689-00 | 68     | Audience A |
| 987.654.709-11 | 70     | Default    |
| 000.000.009-99 | 09     | Audience A |

---

## 🏆 Vexilla technical differentiator

This pattern enables Vexilla to provide:

* **Deterministic Rollouts**
* **Offline-first** evaluation
* Feature flag decisions **100% local**
* Predictable and reproducible rollouts
* Better debuggability

> Vexilla does not rely on randomness or network round-trips to decide flag states.

---

## 📌 When to use this pattern

* Critical feature flags
* Edge / mobile / embedded environments
* Latency-sensitive systems
* Predictable gradual rollouts
* When reproducing user-specific behavior matters

---

## 📎 Notes

This pattern can be applied to any deterministic identifier:

* User ID
* Account ID
* Document number
* Stable hashes

As long as the final output is a **simple numeric bucket**.
