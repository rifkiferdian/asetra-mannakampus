(function () {
    document.querySelectorAll("[data-edit-variant]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("variantEdit");
            ["id", "item", "brand", "code", "model", "sku", "specification", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("variantEdit");
        });
    });
})();
