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
        type: "POST",
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
        type: "POST",
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
        type: "POST",
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

var MenuID = 1000000;
$("#menu-add").click(function() {
    $("#menu ul").append('<li id="menu-item-' + MenuID + '"><div class="collapsible-header menu-item-header">New Item</div><div class="collapsible-body"><span><div class="row"><div class="input-field col s12"> <input id="menu-item-name-' + MenuID + '" class="menu-item-name" type="text"> <label for="menu-item-name-' + MenuID + '">Name</label></div><div class="input-field col s12 hide-on-small-only"> <input id="menu-item-description-regular-' + MenuID + '" class="menu-item-description-regular" type="text"> <label for="menu-item-description-regular-' + MenuID + '">Description</label></div><div class="input-field col s12 hide-on-med-and-up"><textarea id="menu-item-description-large-' + MenuID + '" class="materialize-textarea menu-item-description-large"></textarea><label for="menu-item-description-large-' + MenuID + '">Description</label></div><div class="input-field col s12"> <input id="menu-item-price-' + MenuID + '" class="menu-item-price" type="text"> <label for="menu-item-price-' + MenuID + '">Price</label></div><div class="input-field col"> <a class="btn waves-effect waves-light menu-item-new" onclick="MenuNew(' + MenuID + ');">Submit<i class="material-icons right">send</i></a> <a class="btn waves-effect waves-light red menu-item-delete" onclick="MenuDeleteNew(' + MenuID + ');">Delete<i class="material-icons right">delete</i></a></div></div> </span></div></li>');
    $('.collapsible').collapsible('open', $('#menu ul li').length - 1);
    MenuID++;
});

function ContactDelete(ID) {
    Materialize.toast('Sending delete request!', 4000);

    var contactMessage = "#contact-message-" + ID;

    $.ajax({
        url: "/admin/contact-us/delete",
        type: "POST",
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

function SlideNew(ID) {
    Materialize.toast('Sending new slide request!', 4000);
    var slide = "#slide-" + ID;
    var formData = new FormData();

    if (typeof $(slide + " .slide-image")[0].files[0] === 'undefined') {
        Materialize.Toast.removeAll(); // Clear all other toasts.
        Materialize.toast('You need to select an image to add a new slide.', 4000);
        return;
    }

    formData.append("title", $(slide + " .slide-title").val());
    formData.append("description", $(slide + " .slide-description").val());
    formData.append("imageFile", $(slide + " .slide-image")[0].files[0], $(slide + " .slide-image")[0].files[0].name);
    formData.append("csrfSecret", CsrfSecret);

    $.ajax({
        type: "POST",
        url: "/admin/slide/new",
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function(rRaw) {
            var r = JSON.parse(rRaw);
            if (r.success) {
                $(slide + " .slide-new").attr("onclick", "SlideUpdate(" + r.id + ");");
                $(slide + " .slide-delete").attr("onclick", "SlideDelete(" + r.id + ");");
                $(slide + " .slide-header").text($(slide + " .slide-title").val());
                $(slide).attr("id", "slide-" + r.id);
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully added new slide!', 4000);
            }
        }
    });
}

function SlideUpdate(ID) {
    Materialize.toast('Sending update slide request!', 4000);
    var slide = "#slide-" + ID;
    var formData = new FormData();

    formData.append("id", $(slide + " .slide-id").val());
    formData.append("title", $(slide + " .slide-title").val());
    formData.append("description", $(slide + " .slide-description").val());
    if (typeof $(slide + " .slide-image")[0].files[0] !== 'undefined') {
        formData.append("imageFile", $(slide + " .slide-image")[0].files[0], $(slide + " .slide-image")[0].files[0].name);
        formData.append("cImage", "true"); // cImage is short for changeImage and dictates whether the server will update the image we send or not.
    } else {
        formData.append("cImage", "false"); // cImage is short for changeImage and dictates whether the server will update the image we send or not.
    }
    formData.append("csrfSecret", CsrfSecret);

    $.ajax({
        type: "POST",
        url: "/admin/slide/update",
        data: formData,
        cache: false,
        contentType: false,
        processData: false,
        success: function(rRaw) {
            var r = JSON.parse(rRaw);
            if (r.success) {
                $(slide + " .slide-header").text($(slide + " .slide-title").val());
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully updated slide!', 4000);
            }
        }
    });
}

function SlideDelete(ID) {
    Materialize.toast('Sending delete request!', 4000);

    var slide = "#slide-" + ID;

    $.ajax({
        url: "/admin/slide/delete",
        type: "POST",
        contentType: "application/json; charset=utf-8",
        data: JSON.stringify({
            CsrfSecret: CsrfSecret,
            ID: ID
        }),
        dataType: "json",
        success: function(r) {
            if (r.success) {
                $(slide).remove();
                Materialize.Toast.removeAll(); // Clear all other toasts.
                Materialize.toast('Successfully deleted item!', 4000);
            }
        }
    });
}

function SlideDeleteNew(ID) {
    var slide = "#slide-" + ID;

    $(slide).remove();
    Materialize.toast('Successfully deleted item!', 4000);
}

var SlideID = 1000000;
$("#slide-add").click(function() {
    $("#gallery ul").append('<li id="slide-' + SlideID + '"><div class="collapsible-header slide-header">New Slide</div><div class="collapsible-body"><span><div class="row"> <input hidden value="' + SlideID + '" class="slide-id"/><div class="input-field col s12"> <input id="slide-title-' + SlideID + '" class="slide-title" type="text" name="title"> <label for="slide-title-' + SlideID + '">Title</label></div><div class="input-field col s12"> <input id="slide-description-' + SlideID + '" class="slide-description" type="text" name="description"> <label for="slide-description-' + SlideID + '">Description</label></div><div class="input-field col s12"> <a class="btn waves-effect waves-light slide-image-button" onclick="ImageDialogue(' + SlideID + ');">Add Image<i class="material-icons right">add_a_photo</i></a> <input hidden class="slide-image" type="file"></div><div class="input-field col"> <a class="btn waves-effect waves-light slide-new" onclick="SlideNew(' + SlideID + ');">Submit<i class="material-icons right">send</i></a> <a class="btn waves-effect waves-light red slide-delete" onclick="SlideDeleteNew(' + SlideID + ');">Delete<i class="material-icons right">delete</i></a></div></div> </span></div></li>');
    $('.collapsible').collapsible('open', $('#gallery ul li').length - 1);
    SlideID++;
});

function ImageDialogue(ID) {
    $("#slide-" + ID + " .slide-image").trigger('click');
}