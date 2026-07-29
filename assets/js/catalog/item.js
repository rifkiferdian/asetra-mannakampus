(function () {
    function syncAssetType(form) {
        var candidate = form.querySelector("[data-asset-candidate]");
        var field = form.querySelector("[data-asset-type-field]");
        var select = form.querySelector("[data-asset-type]");
        var enabled = candidate.value === "1";
        field.classList.toggle("hidden", !enabled);
        select.required = enabled;
        if (!enabled) select.value = "";
    }

    document.querySelectorAll("[data-item-form]").forEach(function (form) {
        var candidate = form.querySelector("[data-asset-candidate]");
        candidate.addEventListener("change", function () {
            syncAssetType(form);
        });
    });

    document.querySelectorAll("[data-open]").forEach(function (button) {
        if (button.dataset.open !== "itemCreate") return;
        button.addEventListener("click", function () {
            syncAssetType(document.querySelector("#itemCreate [data-item-form]"));
        });
    });

    document.querySelectorAll("[data-edit-item]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("itemEdit");
            var values = {
                id: button.dataset.id,
                code: button.dataset.code,
                name: button.dataset.name,
                category: button.dataset.category,
                type: button.dataset.type,
                uom: button.dataset.uom,
                assetCandidate: button.dataset.assetCandidate,
                assetType: button.dataset.assetType,
                description: button.dataset.description,
                active: button.dataset.active
            };
            Object.keys(values).forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = values[field] || "";
            });
            syncAssetType(modal.querySelector("[data-item-form]"));
            window.CatalogUI.openModal("itemEdit");
        });
    });
})();
