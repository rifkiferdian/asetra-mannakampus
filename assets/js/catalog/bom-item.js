(function () {
    function syncVariants(form, preferredValue) {
        var component = form.querySelector("[data-component]");
        var variant = form.querySelector("[data-variant]");
        var itemID = component.value;
        var matched = false;
        Array.prototype.forEach.call(variant.options, function (option) {
            if (!option.value) return;
            var visible = option.dataset.item === itemID;
            option.hidden = !visible;
            option.disabled = !visible;
            if (visible && option.value === preferredValue) matched = true;
        });
        if (matched) {
            variant.value = preferredValue;
        } else {
            variant.value = "";
        }
    }

    document.querySelectorAll("[data-bom-item-form]").forEach(function (form) {
        form.querySelector("[data-component]").addEventListener("change", function () {
            syncVariants(form, "");
        });
        syncVariants(form, form.querySelector("[data-variant]").value);
    });

    document.querySelectorAll("[data-edit-bom-item]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("bomItemEdit");
            ["id", "bom", "component", "vendor", "qty", "required", "order", "notes"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            syncVariants(modal.querySelector("[data-bom-item-form]"), button.dataset.variant || "");
            window.CatalogUI.openModal("bomItemEdit");
        });
    });
})();
