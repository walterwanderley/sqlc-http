loadComponents(document);

htmx.onLoad(function (content) {
    loadComponents(content);
});

function loadComponents(content) {
    const alertList = content.querySelectorAll('.alert')
    const alerts = [...alertList].map(element => new bootstrap.Alert(element))
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
    var msgIcon = 'exclamation-triangle-fill';
    var msgClass = 'warning';
    switch (msg.type) {
        case 'error':            
            msgClass = 'danger';
            break
        case 'info':
            msgClass = 'primary';
            msgIcon = 'info-fill'
            break
        case 'success':
            msgClass = 'success';
            msgIcon = 'check-circle-fill'
            break
    }

    const messageDiv = 
        `<div class="alert alert-` + msgClass + ` d-flex align-items-center" role="alert">
    <svg width="24" height="24" class="bi flex-shrink-0 me-2" role="img" aria-label="` + msg.type + `">
        <use xlink:href="#`+ msgIcon + `" />
    </svg>
    <div>
        ` + msg.text + `
    </div>
    <button type="button" class="btn-close ms-auto" data-bs-dismiss="alert" aria-label="Close"></button>
</div>`;

    var messages = htmx.find('#messages');
    console.log('messages', messages);
    messages.innerHTML = messageDiv;
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