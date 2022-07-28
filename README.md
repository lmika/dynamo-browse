# Audax Toolset

A set of small, terminal based UI (TUI, sometimes called "Rougelike") tools for
administering AWS services.

They were built to make it easy to do quick things with
common AWS services, such as DynamoDB, without having to learn incantations with the CLI or
go to the AWS console itself.  This keeps you focused on your task and saves you from
breaking concentration, especially if you do a lot in the terminal.

## The Toolset

More info about the available tools are available here:

- [dynamo-browse](https://audax.tools/docs/dynamo-browse): Browse DynamoDB tables

## Install

Binary packages can be [download from the release page](https://github.com/lmika/audax/releases/latest).

If you have Go 1.18, you can install using the following command:

```
go install github.com/lmika/audax/cmd/dynamo-browse@v0.0.2
```

## License

Audax toolset is released under the MIT License.