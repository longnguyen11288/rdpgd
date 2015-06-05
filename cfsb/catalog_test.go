package cfsb

import "testing"

func TestCatalog(t *testing.T) {
	c := Catalog{}
	err := c.Fetch()
	if err != nil {
		// FAIL: Could not fetch the catalog from database
		t.Fatalf("Get %v", err)
	} else {
		if len(c.Services) == 0 {
			t.Errorf("The services should not be empty")
		} else {
			firstService := c.Services[0]
			if firstService.Id == "" {
				t.Errorf("The service Id is required")
			}
			if firstService.Name == "" {
				t.Errorf("The service Name is required")
			}
			if firstService.Description == "" {
				t.Errorf(" The service description is required")
			}
			if len(firstService.Plans) == 0 {
				t.Errorf("The plans are required")
			} else {
				firstPlan := firstService.Plans[0]
				if firstPlan.Id == "" {
					t.Errorf("The plan Id is required")
				}
				if firstPlan.Name == "" {
					t.Errorf("The plan name is required")
				}
				if firstPlan.Description == "" {
					t.Errorf("The plan descriotion is required")
				}
			}
		}
	}

}
