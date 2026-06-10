# Race Conditions

## Race Condition 1: GetStock

- Code:
```go
func (s *InventoryService) GetStock(productID string) int {
    product := s.products[productID]
    if product == nil {
        return 0
    }
    return product.Stock
}
```
- What happens: reads `Stock` with no lock while another goroutine can be writing it in `Reserve`. Data race on the int field.
- Production scenario:
```go
go svc.Reserve("product", 7)

stock := svc.GetStock("product") 
```
  Might get wrong number, or `-race` complains. In a dashboard showing live stock this would flicker or lie.
- Fix approach: `RLock` around the read. Writers take `Lock` so they don't run together with readers.

## Race Condition 2: Reserve

- Code:
```go
if product.Stock < quantity {
    return ErrInsufficientStock
}
product.Stock -= quantity
```
- What happens: check and subtract are two steps. Two goroutines can both see enough stock and both subtract.
- Production scenario:
```go
// 1 item left in Inventory
go svc.Reserve("phone", 1)
go svc.Reserve("phone", 1)
```
  Both pass `Stock < quantity`, both subtract. You sold 2 phones, had 1. Classic oversell.
- Fix approach: one `Lock`, do check + subtract before `Unlock`.

## Race Condition 3: ReserveMultiple

- Code: first loop checks all items, second loop subtracts. Nothing holds the map between them.
- What happens: another goroutine can change stock after checks passed but before subtract. Or you partially apply, checked A and B ok, something fails mid-way... actually here both loops run without lock so even concurrent `Reserve` on A breaks the assumption from the first loop.
- Production scenario:
```go
// A has 10
go svc.ReserveMultiple([]ReserveItem{{"A", 8}, {"B", 2}})
go svc.Reserve("A", 5)
```
  First goroutine checks A=10 ok, second takes 5 from A, first still subtracts 8. A ends up negative or you get inconsistent order state.
- Fix approach: single write lock for whole method. Validate everything, if ok then subtract everything, if not return without touching stock.

## Race Condition 4: SafeReserve

- Code:
```go
var mu sync.Mutex
mu.Lock()
```
- What happens: new mutex every call. Goroutines don't share it so they don't block each other. Mutex does nothing useful.
- Production scenario:
```go
go svc.SafeReserve("item", 1)
go svc.SafeReserve("item", 1)
```
  Each locks its own `mu`, both modify `Stock` at once. Same race as plain `Reserve`.
- Fix approach: `mu` as struct field, one instance for the whole service.
