ext.related_items("business-addresses", func(item) {
    print("Hello")
    return [
        {"label": "Customer", "query": `city="Austin"`, "args": {"foo": "foo"}},
        {"label": "Payment", "query": `officeOpened=false`, "args": {"daa": "Hello"}},
        {"label": "Thing", "query": `colors.door^="P"`, "args": {"daa": "Hello"}},
    ]
})