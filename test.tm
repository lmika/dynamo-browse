ext.related_items("test-table", func(item) {
    print("Hello")
    return [
        {"label": "Customer", "query": "pk=$foo", "args": {"foo": "foo"}},
        {"label": "Payment", "query": "fla=$daa", "args": {"daa": "Hello"}},
    ]
})