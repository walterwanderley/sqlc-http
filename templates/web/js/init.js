loadComponents(document);

htmx.onLoad(function (content) {
    loadComponents(content);
});

function loadComponents(content) {
    $(content).ready(function () {
        $('.sidenav').sidenav();
    });

    $(".dropdown-trigger").dropdown();

    $('.materialert .close-alert').click(function () {
        $(this).parent().hide('slow');
    });

    $(content).ready(function () {
        $('select').formSelect();
    });

    $(content).ready(function () {
        $('.datepicker').datepicker({
            format: 'dd/mm/yyyy'
        });
    });

    $(content).ready(function () {
        M.updateTextFields();
    });
}

function replacePathParams(event) {
    let pathWithParameters = event.detail.path.replace(/{([A-Za-z0-9_]+)}/g, function (_match, parameterName) {
        let parameterValue = event.detail.parameters[parameterName]
        delete event.detail.parameters[parameterName]
        return parameterValue
    })
    event.detail.path = pathWithParameters
}

function showMessage(msg) {
    var msgIcon = 'check_circle'
    switch (msg.type) {
        case 'error':
            msgIcon = 'error_outline'
            break
        case 'warning':
            msgIcon = 'warning'
            break
        case 'info':
            msgIcon = 'info_outline'
            break
        case 'success':
            msgIcon = 'check'
            break
    }
    const messageDiv =
        `<div class="materialert ` + msg.type + `">
            <div class="material-icons">` + msgIcon + `</div>
            <span>` + msg.text + `</span>
            <button type="button" class="close-alert">×</button>
        </div>`
    var messages = htmx.find('#messages');
    console.log('messages', messages);
    messages.innerHTML = messageDiv;
    $('.materialert .close-alert').click(function () {
        $(this).parent().hide('slow');
    });
}

htmx.on('htmx:responseError', function (evt) {
    try {
        const msg = JSON.parse(evt.detail.xhr.response)
        showMessage(msg)
    } catch (e) {
        const msg = {
            type: 'error',
            text: evt.detail.xhr.response
        }
        showMessage(msg)
    }
});

htmx.on('htmx:sendError', function () {
    const msg = {
        type: 'warning',
        text: 'Server unavailable. Try again in a few minutes.'
    }
    showMessage(msg)
});