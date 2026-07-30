(function () {
    function filterChoices(select, itemID, selectedValue) {
        var selectedIsAvailable = false;

        Array.prototype.forEach.call(select.options, function (option) {
            if (!option.value) {
                return;
            }

            var isAvailable = option.dataset.item === itemID;
            option.hidden = !isAvailable;
            option.disabled = !isAvailable;
            if (isAvailable && option.value === selectedValue) {
                selectedIsAvailable = true;
            }
        });

        select.value = selectedIsAvailable ? selectedValue : "";
    }

    function syncProductChoice(form, selectedBOM, selectedVariant, selectedVendor) {
        var itemSelect = form.querySelector("[data-package-item]");
        var selectedItem = itemSelect.options[itemSelect.selectedIndex];
        var hasItem = Boolean(itemSelect.value);
        var isComposite = hasItem && selectedItem.dataset.type === "COMPOSITE";
        var bomField = form.querySelector("[data-bom-field]");
        var variantField = form.querySelector("[data-variant-field]");
        var vendorField = form.querySelector("[data-vendor-field]");
        var bomSelect = form.querySelector("[data-bom]");
        var variantSelect = form.querySelector("[data-variant]");
        var vendorSelect = form.querySelector("[data-vendor]");

        bomField.classList.toggle("hidden", !isComposite);
        variantField.classList.toggle("hidden", !hasItem || isComposite);
        vendorField.classList.toggle("hidden", !hasItem || isComposite);

        bomSelect.required = isComposite;
        variantSelect.required = hasItem && !isComposite;
        vendorSelect.required = hasItem && !isComposite;

        if (isComposite) {
            filterChoices(bomSelect, itemSelect.value, selectedBOM || "");
            variantSelect.value = "";
            vendorSelect.value = "";
            return;
        }

        bomSelect.value = "";
        filterChoices(variantSelect, itemSelect.value, selectedVariant || "");
        vendorSelect.value = selectedVendor || "";
    }

    document.querySelectorAll("[data-package-item-form]").forEach(function (form) {
        var itemSelect = form.querySelector("[data-package-item]");
        syncProductChoice(form, "", "", "");
        itemSelect.addEventListener("change", function () {
            syncProductChoice(form, "", "", "");
        });
    });

    document.querySelectorAll("[data-edit-package-item]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("packageItemEdit");
            ["id", "package", "item", "qty", "optional", "order", "notes"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            syncProductChoice(
                modal,
                button.dataset.bom || "",
                button.dataset.variant || "",
                button.dataset.vendor || ""
            );
            window.CatalogUI.openModal("packageItemEdit");
        });
    });
})();
