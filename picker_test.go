package picker

import (
	"math/big"
	"os"
	"testing"
)

type Posting struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type VoucherBody struct {
	ID                     int64     `json:"id"`
	Number                 int       `json:"number"`
	Date                   string    `json:"date"`
	Description            string    `json:"description"`
	NumberAsString         string    `json:"numberAsString"`
	ExternalVoucherNumber  string    `json:"externalVoucherNumber"`
	TempNumber             int       `json:"tempNumber"`
	Version                int       `json:"version"`
	Year                   int       `json:"year"`
	WasAutoMatched         bool      `json:"wasAutoMatched"`
	URL                    string    `json:"url"`
	Postings               []Posting `json:"postings"`
	VendorInvoiceNumber    *string   `json:"vendorInvoiceNumber"`
	Attachment             *string   `json:"attachment"`
	Document               *string   `json:"document"`
	EdiDocument            *string   `json:"ediDocument"`
	ReverseVoucher         *string   `json:"reverseVoucher"`
	SupplierVoucherType    *string   `json:"supplierVoucherType"`
	VoucherType            *string   `json:"voucherType"`
}

type WebhookData struct {
	SubscriptionID       int64     `json:"subscriptionId"`
	Event                string    `json:"event"`
	ObjectID             int64     `json:"objectId"`
	CompanyDataSourceID  int       `json:"companyDataSourceId"`
	Body                 VoucherBody `json:"body"`
	ReceivedAt           string    `json:"receivedAt"`
}

func TestPickToStruct(t *testing.T) {
	// Read the JSON file
	jsonData, err := os.ReadFile("../voucher-demo.json")
	if err != nil {
		t.Fatalf("Failed to read voucher-demo.json: %v", err)
	}

	// Parse into struct using PickToStruct
	var webhook WebhookData
	err = PickToStruct(string(jsonData), &webhook)
	if err != nil {
		t.Fatalf("PickToStruct failed: %v", err)
	}

	// Verify the parsed data
	if webhook.Event != "voucher.update" {
		t.Errorf("Expected event 'voucher.update', got '%s'", webhook.Event)
	}

	if webhook.ObjectID != 567396756 {
		t.Errorf("Expected objectId 567396756, got %d", webhook.ObjectID)
	}

	if webhook.SubscriptionID != 18652258 {
		t.Errorf("Expected subscriptionId 18652258, got %d", webhook.SubscriptionID)
	}

	if webhook.Body.ID != 567396756 {
		t.Errorf("Expected body.id 567396756, got %d", webhook.Body.ID)
	}

	if webhook.Body.Description != "Test 1" {
		t.Errorf("Expected body.description 'Test 1', got '%s'", webhook.Body.Description)
	}

	if webhook.Body.Date != "2025-10-30" {
		t.Errorf("Expected body.date '2025-10-30', got '%s'", webhook.Body.Date)
	}

	if len(webhook.Body.Postings) != 3 {
		t.Errorf("Expected 3 postings, got %d", len(webhook.Body.Postings))
	}

	if len(webhook.Body.Postings) > 0 {
		if webhook.Body.Postings[0].ID != 3623299565 {
			t.Errorf("Expected first posting id 3623299565, got %d", webhook.Body.Postings[0].ID)
		}
		t.Logf("First posting: %+v", webhook.Body.Postings[0])
	} else {
		t.Log("No postings found - this might be expected if slice parsing isn't implemented yet")
	}

	if webhook.Body.WasAutoMatched != false {
		t.Errorf("Expected wasAutoMatched false, got %t", webhook.Body.WasAutoMatched)
	}

	t.Logf("Successfully parsed webhook data: %+v", webhook)
	t.Logf("Body: %+v", webhook.Body)
}

func TestPickToStructWithPrimitiveSlices(t *testing.T) {
	type TestStruct struct {
		Names   []string  `json:"names"`
		Numbers []float64 `json:"numbers"`
		Floats  []float64 `json:"floats"`
		Flags   []bool    `json:"flags"`
	}
	
	jsonData := `{
		"names": ["Alice", "Bob", "Charlie"],
		"numbers": [1.0, 2.0, 3.0, 4.0, 5.0],
		"floats": [1.1, 2.2, 3.3],
		"flags": [true, false, true]
	}`
	
	var result TestStruct
	err := PickToStruct(jsonData, &result)
	if err != nil {
		t.Fatalf("PickToStruct failed: %v", err)
	}
	
	// Verify string slice
	if len(result.Names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(result.Names))
	}
	if result.Names[0] != "Alice" {
		t.Errorf("Expected first name 'Alice', got '%s'", result.Names[0])
	}
	
	// Verify number slice (float64)
	if len(result.Numbers) != 5 {
		t.Errorf("Expected 5 numbers, got %d", len(result.Numbers))
	}
	if result.Numbers[0] != 1.0 {
		t.Errorf("Expected first number 1.0, got %f", result.Numbers[0])
	}
	
	// Verify float slice
	if len(result.Floats) != 3 {
		t.Errorf("Expected 3 floats, got %d", len(result.Floats))
	}
	if result.Floats[0] != 1.1 {
		t.Errorf("Expected first float 1.1, got %f", result.Floats[0])
	}
	
	// Verify bool slice
	if len(result.Flags) != 3 {
		t.Errorf("Expected 3 flags, got %d", len(result.Flags))
	}
	if result.Flags[0] != true {
		t.Errorf("Expected first flag true, got %t", result.Flags[0])
	}
	
	t.Logf("Successfully parsed primitive slices: %+v", result)
}

func TestGetTypedArray(t *testing.T) {
	testData := map[string]interface{}{
		"strings":    []interface{}{"hello", "world", "test"},
		"ints":       []interface{}{int64(1), int64(2), int64(3)},
		"floats":     []interface{}{1.1, 2.2, 3.3},
		"bools":      []interface{}{true, false, true},
		"bigints":    []interface{}{big.NewInt(100), big.NewInt(200)},
		"bigfloats":  []interface{}{big.NewFloat(1.23), big.NewFloat(4.56)},
		"bigrats":    []interface{}{big.NewRat(1, 2), big.NewRat(3, 4)},
		"mixed":      []interface{}{"string", int64(123)}, // Should cause error
		"notarray":   "this is not an array",
	}
	
	picker := NewPicker(testData)
	
	// Test successful string array conversion
	result := picker.GetTypedArray("strings", ValueTypeString)
	if result == nil {
		t.Error("Expected non-nil result for strings")
	} else {
		stringArray := result.([]string)
		if len(stringArray) != 3 {
			t.Errorf("Expected 3 strings, got %d", len(stringArray))
		}
		if stringArray[0] != "hello" {
			t.Errorf("Expected first string 'hello', got '%s'", stringArray[0])
		}
	}
	
	// Test successful int array conversion
	result = picker.GetTypedArray("ints", ValueTypeInt)
	if result == nil {
		t.Error("Expected non-nil result for ints")
	} else {
		intArray := result.([]int64)
		if len(intArray) != 3 {
			t.Errorf("Expected 3 ints, got %d", len(intArray))
		}
		if intArray[0] != 1 {
			t.Errorf("Expected first int 1, got %d", intArray[0])
		}
	}
	
	// Test successful float array conversion
	result = picker.GetTypedArray("floats", ValueTypeFloat)
	if result == nil {
		t.Error("Expected non-nil result for floats")
	} else {
		floatArray := result.([]float64)
		if len(floatArray) != 3 {
			t.Errorf("Expected 3 floats, got %d", len(floatArray))
		}
		if floatArray[0] != 1.1 {
			t.Errorf("Expected first float 1.1, got %f", floatArray[0])
		}
	}
	
	// Test successful bool array conversion
	result = picker.GetTypedArray("bools", ValueTypeBool)
	if result == nil {
		t.Error("Expected non-nil result for bools")
	} else {
		boolArray := result.([]bool)
		if len(boolArray) != 3 {
			t.Errorf("Expected 3 bools, got %d", len(boolArray))
		}
		if boolArray[0] != true {
			t.Errorf("Expected first bool true, got %t", boolArray[0])
		}
	}
	
	// Test successful big.Int array conversion
	result = picker.GetTypedArray("bigints", ValueTypeBigInt)
	if result == nil {
		t.Error("Expected non-nil result for bigints")
	} else {
		bigintArray := result.([]*big.Int)
		if len(bigintArray) != 2 {
			t.Errorf("Expected 2 big ints, got %d", len(bigintArray))
		}
		if bigintArray[0].Int64() != 100 {
			t.Errorf("Expected first big int 100, got %d", bigintArray[0].Int64())
		}
	}
	
	// Test error cases
	picker.GetTypedArray("mixed", ValueTypeString) // Should add error
	picker.GetTypedArray("notarray", ValueTypeString) // Should add error
	picker.GetTypedArray("nonexistent", ValueTypeString) // Should add error
	
	if !picker.HasErrors() {
		t.Error("Expected picker to have errors after invalid operations")
	}
	
	errorKeys := picker.ErrorKeys()
	if len(errorKeys) != 3 {
		t.Errorf("Expected 3 error keys, got %d", len(errorKeys))
	}
}

func TestNested(t *testing.T) {
	testData := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"age":  int64(30),
			"profile": map[string]interface{}{
				"email": "john@example.com",
			},
		},
		"notobject": "this is not an object",
	}
	
	picker := NewPicker(testData)
	
	// Test successful nested access
	userPicker := picker.Nested("user")
	name := userPicker.GetString("name")
	if name != "John" {
		t.Errorf("Expected name 'John', got '%s'", name)
	}
	
	// Test nested in nested
	profilePicker := userPicker.Nested("profile")
	email := profilePicker.GetString("email")
	if email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", email)
	}
	
	// Test error case
	picker.Nested("notobject")
	picker.Nested("nonexistent")
	
	if !picker.HasErrors() {
		t.Error("Expected picker to have errors after invalid nested operations")
	}
}

func TestNestedArray(t *testing.T) {
	testData := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "John",
				"age":  int64(30),
			},
			map[string]interface{}{
				"name": "Jane",
				"age":  int64(25),
			},
		},
		"notarray": "this is not an array",
		"invalidarray": []interface{}{
			"not an object",
			123,
		},
	}
	
	picker := NewPicker(testData)
	
	// Test successful nested array access
	usersArray := picker.NestedArray("users")
	if len(usersArray.Items) != 2 {
		t.Errorf("Expected 2 users, got %d", len(usersArray.Items))
	}
	
	// Access individual items
	firstUser := usersArray.Items[0]
	name := firstUser.GetString("name")
	if name != "John" {
		t.Errorf("Expected first user name 'John', got '%s'", name)
	}
	
	secondUser := usersArray.Items[1]
	name = secondUser.GetString("name")
	if name != "Jane" {
		t.Errorf("Expected second user name 'Jane', got '%s'", name)
	}
	
	// Test error cases
	picker.NestedArray("notarray")
	picker.NestedArray("nonexistent")
	picker.NestedArray("invalidarray") // Contains non-objects
	
	if !picker.HasErrors() {
		t.Error("Expected picker to have errors after invalid nested array operations")
	}
}

func TestConfirm(t *testing.T) {
	// Test successful confirm
	testData := map[string]interface{}{
		"name": "John",
		"age":  int64(30),
	}
	
	picker := NewPicker(testData)
	picker.GetString("name") // Valid operation
	
	err := picker.Confirm()
	if err != nil {
		t.Errorf("Expected no error from Confirm(), got: %v", err)
	}
	
	// Test confirm with errors
	picker.GetString("nonexistent") // Should add error
	picker.GetInt("name") // Should add error (wrong type)
	
	err = picker.Confirm()
	if err == nil {
		t.Error("Expected error from Confirm() when picker has errors")
	}
	
	expectedKeys := []string{"nonexistent", "name"}
	for _, key := range expectedKeys {
		found := false
		for _, errorKey := range picker.ErrorKeys() {
			if errorKey == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error key '%s' not found in error keys", key)
		}
	}
	
	// Test confirm on nested picker (should fail)
	testDataNested := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
		},
	}
	
	pickerNested := NewPicker(testDataNested)
	nestedPicker := pickerNested.Nested("user")
	
	err = nestedPicker.Confirm()
	if err == nil {
		t.Error("Expected error when calling Confirm() on nested picker")
	}
}