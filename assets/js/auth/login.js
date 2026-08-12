'use strict';

document.addEventListener('DOMContentLoaded', function () {
    var toggle = document.querySelector('[data-password-toggle]');
    var password = document.getElementById('password');

    if (!toggle || !password) {
        return;
    }

    var icon = toggle.querySelector('i');

    toggle.addEventListener('click', function () {
        var shouldShowPassword = password.type === 'password';

        password.type = shouldShowPassword ? 'text' : 'password';
        toggle.setAttribute('aria-label', shouldShowPassword ? 'Sembunyikan password' : 'Tampilkan password');
        toggle.setAttribute('aria-pressed', String(shouldShowPassword));

        if (icon) {
            icon.classList.toggle('bx-show', !shouldShowPassword);
            icon.classList.toggle('bx-hide', shouldShowPassword);
        }
    });
});

