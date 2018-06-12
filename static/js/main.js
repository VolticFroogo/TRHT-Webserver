$(document).ready(function(){
    // Materialize initializations
    Waves.displayEffect();
    $(".button-collapse").sideNav();
    $('.slider').slider();
    $('.parallax').parallax();
    $('select').material_select();

    // Smooth scrolling and navbar compensation
    $(document).on('click', 'a[href^="#"]', function (event) {
        event.preventDefault();
    
        $('html, body').animate({
            scrollTop: $($.attr(this, 'href')).offset().top - 75
        }, 500);
    });
});

$("#contact-us .btn").click(function(){
    var d = new Date(); // Get current date.
    var epochMilliseconds = d.getTime(); // Get milliseconds since epoch.

    localStorage = window.localStorage; // Set localStorage to localStorage.
    nextMessage = localStorage.getItem("nextContactUsMessage"); // Get next message time.
    if (nextMessage < epochMilliseconds || nextMessage === null) {
        // If it's been 60 seconds since the last message.
        localStorage.setItem("nextContactUsMessage", epochMilliseconds + 60000); // Set the next message time to the current time + 60 seconds.
        Materialize.toast('Sending message!', 4000);

        $.ajax({
            url: "/contact-us",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                Name: $("#contact-us .name").val(),
                Email: $("#contact-us .email").val(),
                Message: $("#contact-us .message").val(),
                Captcha: grecaptcha.getResponse()
            }),
            dataType: "json",
            success: function(r) {
                if (r.success) {
                    Materialize.Toast.removeAll(); // Clear all other toasts.
                    Materialize.toast('Successfully sent message!', 4000);
                }
            }
        });
        
        grecaptcha.reset(); // Reset the recaptcha
    } else {
        Materialize.toast('Please wait 1 minute between sending each message.', 4000);
    }
});