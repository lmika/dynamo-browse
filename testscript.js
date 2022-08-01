const dynamobrowse = require("audax:dynamo-browse");

dynamobrowse.session.registerCommand("bla", () => {
    dynamobrowse.ui.prompt("What is your name? ").then((name) => {
        dynamobrowse.ui.alert("Hello, " + name);
    })
})

