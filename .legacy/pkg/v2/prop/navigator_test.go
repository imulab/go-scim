package prop

func ExampleNavigator() {
	getResource := func() *Resource {
		return &Resource{}
	}
	// create navigator
	nav := getResource().Navigator()
	// traverse resource structure
	nav.Dot("emails").At(1).Dot("value")
	// check for errors during the chained traversal
	if err := nav.Error(); err != nil {
		panic(err)
	}
	// access the property at the top of the trace stack
	println(nav.Current().Raw())
}
