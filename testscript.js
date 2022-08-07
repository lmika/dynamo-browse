const dynamobrowse = require("audax:dynamo-browse");

dynamobrowse.session.registerCommand("bla", () => {
    dynamobrowse.session.query(`pk^="02"`, {
        table: "inventorysadasdas"
    }).then((rs) => {
        dynamobrowse.session.currentResultSet = rs;
        dynamobrowse.ui.alert("Length: " + rs.rows.length);
    }).catch((err) => {
        dynamobrowse.ui.alert("Error: " + err);
    });

    // let rs = dynamobrowse.session.currentResultSet;
    // let tableName = rs.table.name;
    //
    // rs.rows[0].item.address = "123 Fake St.";
    //
    // dynamobrowse.ui.alert("Len rows = " + rs.rows.length);
    // dynamobrowse.ui.alert("PK = " + rs.rows[0].item.pk);
    //
    //
    //dynamobrowse.ui.alert("Current table name = " + tableName);
    // dynamobrowse.ui.prompt("What is your name? ").then((name) => {
    //     dynamobrowse.ui.alert("Hello, " + name);
    // });
})

