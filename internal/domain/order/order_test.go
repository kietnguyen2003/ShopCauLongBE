package order

import "testing"

func TestAddItemRejectsModificationAfterConfirmation(t *testing.T) {
	ord, err := NewOrder(1, "Test User", "0123456789", "HCM", "user@example.com")
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	snapshot, err := NewProductSnapshot("Yonex Shirt", 100000, "img", "clothing", "desc")
	if err != nil {
		t.Fatalf("NewProductSnapshot() error = %v", err)
	}

	if err := ord.AddItem(1, snapshot, 1); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	if err := ord.UpdateStatus(OrderStatusConfirmed); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if err := ord.AddItem(2, snapshot, 1); err != ErrOrderCannotBeModified {
		t.Fatalf("expected ErrOrderCannotBeModified, got %v", err)
	}
}

func TestUpdateStatusRejectsInvalidTransition(t *testing.T) {
	ord, err := NewOrder(1, "Test User", "0123456789", "HCM", "user@example.com")
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	if err := ord.UpdateStatus(OrderStatusConfirmed); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	if err := ord.UpdateStatus(OrderStatusCancelled); err != ErrInvalidOrderStatus {
		t.Fatalf("expected ErrInvalidOrderStatus, got %v", err)
	}
}

func TestValidateForCreationRequiresItems(t *testing.T) {
	ord, err := NewOrder(1, "Test User", "0123456789", "HCM", "user@example.com")
	if err != nil {
		t.Fatalf("NewOrder() error = %v", err)
	}

	if err := ord.ValidateForCreation(); err != ErrOrderMustHaveItems {
		t.Fatalf("expected ErrOrderMustHaveItems, got %v", err)
	}
}
