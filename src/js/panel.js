let products, pr_id, users, usr_id;
$(function() {
    $("button").click(function () {
        switch($(this).attr("id")){
            case "register":
                $("#panel").hide();
                $("#btnExit").hide();
                $("#reg").show();
                $("#btnBack").show();
                $("#heading").text("Preču reģistrēšana");
                break;
            case "admin":
                $.get( "/user", function(data) {
                    users = JSON.parse(data);
                    users.forEach((element, index) => {
                        $('#adminTable tr:last').after(' <tr>\n' +
                            '            <td>'+element.Id+'</td>\n' +
                            '            <td>'+element.FName+'</td>\n' +
                            '            <td>'+element.LName+'</td>\n' +
                            '            <td>'+element.Role+'</td>\n' +
                            '            <td class="btn-sm">\n' +
                            '                <button class="editBtn" onclick="admBut('+index+')">edit</button>\n' +
                            '             <button class="removeBtn" onclick="removeBut('+index+')" style="background: #f44336">-</button>\n'+
                            '            </td>\n' +
                            '        </tr>');
                    });
                    $('#adminTable tr:last').after(' <tr><td colspan="5"><button id="createBut" style="background: #4CAF50; width: 100%; height: 40px; color: white; cursor:pointer; font-size: 30px" onclick="createUser()">Izveidot jaunu lietotāju</button></td></tr>');

                });
                $("#panel").hide();
                $("#btnExit").hide();
                $("#adminPage").show();
                $("#btnBack").show();
                $("#heading").text("Administrēšanas panelis");
                break;
            case "work":
                $("#panel").hide();
                $("#btnExit").hide();
                $("#workPage").show();
                $("#btnBack").show();
                $("#heading").text("Darba vide");
                break;
            case "btnBack":
                location.reload();
                break;
            case "SubmitReg":
                $.post( "/product", {name: $("#Name").val(), qnty: $("#Quantity").val(), date: $("#Date").val(), before: $("#DateTo").val()})
                    .done(function() {
                        location.reload();
                    });
                break;
            case "sell":
                $.get( "/product", function(data) {
                    products = JSON.parse(data);
                    let textproducts = [];
                    products.forEach((element) => {
                        textproducts.push(element.Name+"  |>>>| Daudzums: "+element.Qnty)
                    })
                    Swal.fire({
                        title: 'Izvēlēties produktu',
                        input: 'select',
                        inputOptions: {
                            'Produkti': textproducts,
                        },
                        inputPlaceholder: 'Izvēlēties produktu',
                        showCancelButton: true,
                    }).then((result) => {
                        if (result.value) {
                            $.ajax({
                                type: 'PATCH',
                                url: '/product',
                                data: JSON.stringify({Id: products[result.value].Id}),
                                processData: false,
                            });

                        }
                    })
                });
                break;
            case "edit":
                $.get( "/product", function(data) {
                    products = JSON.parse(data);
                    products.forEach((element) => {
                        $('#editTable tr:last').after(' <tr>\n' +
                            '            <td>'+element.Name+'</td>\n' +
                            '            <td>'+element.Date+'</td>\n' +
                            '            <td>'+element.Expires+'</td>\n' +
                            '            <td>'+element.Qnty+'</td>\n' +
                            '            <td class="btn-sm">\n' +
                            '                <button class="editBtn" onclick="editBut('+element.Id+')">+</button>\n' +
                            '            </td>\n' +
                            '        </tr>');
                    });

                });
                $("#workPage").hide();
                $("#btnExit").hide();
                $("#editPage").show();
                $("#btnBack").show();
                break;
            case "eSubmitReg":
                $.ajax({
                    type: 'PATCH',
                    url: '/product',
                    data: "u"+JSON.stringify({Id: pr_id, Name: $("#eName").val(), Qnty: parseInt($("#eQuantity").val()), Date: $("#eDate").val(), Expires: $("#eDateTo").val()}),
                    processData: false,
                });
                location.reload();
                break;
            case "eSubmitDelete":
                $.ajax({
                    type: 'PATCH',
                    url: '/product',
                    data: "d"+JSON.stringify({Id: pr_id}),
                    processData: false,
                });
                location.reload();
                break;
            default:
                $.get( "/logoff", function() {
                    location.replace("/");
                });
        }
    });
});

function editBut(id) {
    $("#editPage").hide();
    $("#editing").show();
    $("#btnBack").show();
    products.forEach((element) => {
       if (element.Id === id)
        $("#eName").val(element.Name);
        $("#eQuantity").val(element.Qnty);
        $("#eDate").val(element.Date);
        $("#eDateTo").val(element.Expires);
    });
    pr_id = id;
}

function admBut(id) {
    $("#adminPage").hide();
    $("#admining").show();
    $("#btnBack").show();
    Swal.fire({
        title: 'Izvēlēties lomu',
        input: 'select',
        inputOptions: {
            'Lomas': {
                admin : "Administrātors",
                seller : "Pārdēvējs",
            },
        },
        inputPlaceholder: 'Izvēlēties lomu',
        showCancelButton: true,
    }).then((result) => {
        if (result.value) {
            $.ajax({
                type: 'PATCH',
                url: '/user',
                data: JSON.stringify({Id: users[id].Id, Role: result.value}),
                processData: false,
                statusCode: {
                    403: function() {
                        alert( "Jūs nevarat rediģēt savu kontu!" );
                    }
                }
            });
            usr_id = id;
            location.reload();
        }
    });



}

function removeBut(id) {
            $.ajax({
                type: 'DELETE',
                url: '/user',
                data: JSON.stringify({Id: users[id].Id}),
                processData: false,
                statusCode: {
                    403: function() {
                        alert( "Jūs nevarat rediģēt savu kontu!" );
                    }
                }
            });
            usr_id = id;
            location.reload();

}

function createUser(){
    location.replace("src/new.html")
}