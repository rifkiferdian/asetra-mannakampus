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

    function syncItemType(form, resetValue) {
        var itemType = form.querySelector("[data-item-type]");
        var purchasable = form.querySelector("[data-purchasable]");
        var selected = itemType.options[itemType.selectedIndex];
        var composite = selected && selected.dataset.code === "COMPOSITE";
        purchasable.disabled = composite;
        if (composite) {
            purchasable.value = "0";
        } else if (resetValue) {
            purchasable.value = "1";
        }
    }

    document.querySelectorAll("[data-item-form]").forEach(function (form) {
        var candidate = form.querySelector("[data-asset-candidate]");
        var itemType = form.querySelector("[data-item-type]");
        candidate.addEventListener("change", function () {
            syncAssetType(form);
        });
        itemType.addEventListener("change", function () {
            syncItemType(form, true);
        });
    });

    document.querySelectorAll("[data-open]").forEach(function (button) {
        if (button.dataset.open !== "itemCreate") return;
        button.addEventListener("click", function () {
            var form = document.querySelector("#itemCreate [data-item-form]");
            syncAssetType(form);
            syncItemType(form, false);
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
                componentType: button.dataset.componentType,
                purchasable: button.dataset.purchasable,
                description: button.dataset.description,
                active: button.dataset.active
            };
            Object.keys(values).forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = values[field] || "";
            });
            var form = modal.querySelector("[data-item-form]");
            syncAssetType(form);
            syncItemType(form, false);
            window.CatalogUI.openModal("itemEdit");
        });
    });
})();
