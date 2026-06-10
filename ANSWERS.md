# Answers

## Q1

Mutex works only if goroutines lock the same object. Here `mu` is created inside the function, so every call gets a fresh mutex on the stack. Two goroutines calling `SafeReserve` at the same time each lock their own mutex, they don't wait for each other and both go into the critical section together. Basically the same as no lock. Need to put `mu` on the struct so everyone shares one.

## Q2

Deadlock. G1 locks A, wants B. G2 locks B, wants A. Both sit and wait forever.

With per-product locks you need a fixed order, sort product IDs and always lock in that order. Or one lock on the whole service, less moving parts and you don't have to worry about who locks what first.

## Q3

You release the lock after reading stock, then check, then lock again to write. Window in between where another goroutine can also read and pass the check.

Say stock is 1. G1 reads 1, unlocks, check ok. G2 reads 1, unlocks, check ok. Both subtract, stock goes to -1, two reservations for one item.

Without locks it's obviously broken. With this fix it still breaks but looks safe, plus you pay for lock/unlock twice for nothing. Overselling is the bug.

## Q4

No. `-race` only catches what actually happened during that test run. Didn't hit the bad path, no warning. Race might be there on another load pattern or in prod under more traffic.

Also doesn't run in production normally. Clean race detector output is a good sign but not proof.
