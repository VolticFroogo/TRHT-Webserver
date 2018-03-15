$(document).ready(function(){
    // Materialize initializations
    Waves.displayEffect();
    $(".button-collapse").sideNav();
});

function MenuUpdate(ID) {
    Materialize.toast('Sending update request!', 4000);

    var menuItem = "#menu-item-" + ID;
    var descriptionStatement;
    if ($(window).width() <= 600) {      
        descriptionStatement = menuItem + " .menu-item-description-large"; // Get large textbox for small screens.
    } else {
        descriptionStatement = menuItem + " .menu-item-description-regular"; // Get regular textbox for other screens.
    }

    $.ajax({
        url: "/admin/menu",
        type: "post",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            CsrfSecret: CsrfSecret,
            ID: ID,
            Name: $(menuItem + " .menu-item-name").val(),
            Description: $(descriptionStatement).val(),
            Price: $(menuItem + " .menu-item-price").val()
        }),
        dataType: "json",
        success: function(r) {
            if (r.success) {
                $(menuItem + " .menu-item-header").text($(menuItem + " .menu-item-name").val());
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully updated!', 4000);
            }
        }
    });
}

function MenuNew(ID) {
    Materialize.toast('Sending new item request!', 4000);

    var menuItem = "#menu-item-" + ID;
    var descriptionStatement;
    if ($(window).width() <= 600) {      
        descriptionStatement = menuItem + " .menu-item-description-large"; // Get large textbox for small screens.
    } else {
        descriptionStatement = menuItem + " .menu-item-description-regular"; // Get regular textbox for other screens.
    }

    $.ajax({
        url: "/admin/menu/new",
        type: "post",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            CsrfSecret: CsrfSecret,
            Name: $(menuItem + " .menu-item-name").val(),
            Description: $(descriptionStatement).val(),
            Price: $(menuItem + " .menu-item-price").val()
        }),
        dataType: "json",
        success: function(r) {
            if (r.success) {
                $(menuItem + " .menu-item-new").attr("onclick", "MenuUpdate(" + r.id + ");");
                $(menuItem + " .menu-item-delete").attr("onclick", "MenuDelete(" + r.id + ");");
                $(menuItem + " .menu-item-header").text($(menuItem + " .menu-item-name").val());
                $(menuItem).attr("id", "menu-item-" + r.id);
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully added new item!', 4000);
            }
        }
    });
}

function MenuDelete(ID) {
    Materialize.toast('Sending delete request!', 4000);

    var menuItem = "#menu-item-" + ID;

    $.ajax({
        url: "/admin/menu/delete",
        type: "post",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            CsrfSecret: CsrfSecret,
            ID: ID
        }),
        dataType: "json",
        success: function(r) {
            if (r.success) {
                $(menuItem).remove();
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully deleted item!', 4000);
            }
        }
    });
}

function MenuDeleteNew(ID) {
    var menuItem = "#menu-item-" + ID;

    $(menuItem).remove();
    Materialize.toast('Successfully deleted item!', 4000);
}

var ID = 1000000;
$("#menu-add").click(function() {
    ID++;
    $("#menu ul").append('<li id="menu-item-' + ID + '"> <div class="collapsible-header menu-item-header">New Item</div><div class="collapsible-body"><span> <div class="row"> <div class="input-field col s12"> <input id="menu-item-name-' + ID + '" class="menu-item-name" type="text"> <label for="menu-item-name-' + ID + '">Name</label> </div><div class="input-field col s12 hide-on-small-only"> <input id="menu-item-description-regular-' + ID + '" class="menu-item-description-regular" type="text"> <label for="menu-item-description-regular-' + ID + '">Description</label> </div><div class="input-field col s12 hide-on-med-and-up"> <textarea id="menu-item-description-large-' + ID + '" class="materialize-textarea menu-item-description-large"></textarea> <label for="menu-item-description-large-' + ID + '">Description</label> </div><div class="input-field col s12"> <input id="menu-item-price-' + ID + '" class="menu-item-price" type="text"> <label for="menu-item-price-' + ID + '">Price</label> </div><div class="input-field col"> <a class="btn waves-effect waves-light menu-item-new" onclick="MenuNew(' + ID + ');">Submit<i class="material-icons right">send</i></a> <a class="btn waves-effect waves-light red menu-item-delete" onclick="MenuDeleteNew(' + ID + ');">Delete<i class="material-icons right">delete</i></a> </div></div></span></div></li>');
    $('.collapsible').collapsible('open', $('#menu ul li').length - 1);
});

function ContactDelete(ID) {
    Materialize.toast('Sending delete request!', 4000);

    var contactMessage = "#contact-message-" + ID;

    $.ajax({
        url: "/admin/contact-us/delete",
        type: "post",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            CsrfSecret: CsrfSecret,
            ID: ID
        }),
        dataType: "json",
        success: function(r) {
            if (r.success) {
                $(contactMessage).remove();
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully deleted item!', 4000);
            }
        }
    });
}