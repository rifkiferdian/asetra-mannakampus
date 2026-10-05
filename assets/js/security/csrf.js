(function () {
    'use strict';

    function readCookie(name) {
        var prefix = encodeURIComponent(name) + '=';
        var cookies = document.cookie ? document.cookie.split('; ') : [];
        for (var i = 0; i < cookies.length; i += 1) {
            if (cookies[i].indexOf(prefix) === 0) {
                return decodeURIComponent(cookies[i].slice(prefix.length));
            }
        }
        return '';
    }

    function token() {
        return readCookie('asetra_csrf');
    }

    function protectForm(form) {
        if (!form || String(form.method || 'get').toLowerCase() === 'get') return;
        if (new URL(form.action || window.location.href, window.location.href).origin !== window.location.origin) return;
        var value = token();
        if (!value) return;
        var input = form.querySelector('input[name="_csrf"]');
        if (!input) {
            input = document.createElement('input');
            input.type = 'hidden';
            input.name = '_csrf';
            form.appendChild(input);
        }
        input.value = value;
    }

    function protectAllForms() {
        document.querySelectorAll('form').forEach(protectForm);
    }

    window.asetraSubmitPost = function (url) {
        if (!url) return;
        var form = document.createElement('form');
        form.method = 'post';
        form.action = url;
        form.hidden = true;
        document.body.appendChild(form);
        protectForm(form);
        form.submit();
    };

    document.addEventListener('submit', function (event) {
        protectForm(event.target);
    }, true);

    if (window.fetch) {
        var originalFetch = window.fetch;
        window.fetch = function (resource, options) {
            var requestOptions = options || {};
            var method = String(requestOptions.method || (resource && resource.method) || 'GET').toUpperCase();
            var resourceURL = typeof resource === 'string' ? resource : resource.url;
            var isSameOrigin = new URL(resourceURL, window.location.href).origin === window.location.origin;
            if (isSameOrigin && !['GET', 'HEAD', 'OPTIONS', 'TRACE'].includes(method)) {
                var headers = new Headers(requestOptions.headers || {});
                headers.set('X-CSRF-Token', token());
                requestOptions = Object.assign({}, requestOptions, { headers: headers });
            }
            return originalFetch.call(this, resource, requestOptions);
        };
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', protectAllForms);
    } else {
        protectAllForms();
    }
}());
