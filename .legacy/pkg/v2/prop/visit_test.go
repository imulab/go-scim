package prop

func ExampleVisit() {
	getVisitor := func() Visitor {
		// type ExampleVisitor struct {}
		//
		// func (v *ExampleVisitor) ShouldVisit(_ Property) bool {
		// 	return true
		// }
		//
		// func (v *ExampleVisitor) Visit(property Property) error {
		// 	println(property.Raw())
		// 	return nil
		// }
		//
		// func (v *ExampleVisitor) BeginChildren(container Property) {
		// 	println("entering", container.Attribute().Path())
		// }
		//
		// func (v *ExampleVisitor) EndChildren(container Property) {
		// 	println("exiting", container.Attribute().Path())
		// }
		// assuming it returns the above ExampleVisitor
		return nil
	}
	v := getVisitor()

	getProperty := func() Property {
		// Assuming a property structure of:
		//	{
		//		"id": "foobar",
		//		"name": {
		//			"givenName": "David",
		//			"familyName": "Q"
		//		}
		//	}
		// assuming it returns the above property
		return nil
	}
	p := getProperty()

	_ = Visit(p, v)
	// should print:
	//
	// foobar
	// map[string]interface{}{"givenName": "David", "familyName": "Q"}
	// entering name
	// David
	// Q
	// exiting name
}
