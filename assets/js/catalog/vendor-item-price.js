(function () {
    document.querySelectorAll("[data-edit-vendor-price]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("vendorPriceEdit");
            var fields = [
                "id", "vendor", "variant", "price", "currency", "minimum",
                "validFrom", "validUntil", "leadTime", "quotation",
                "preferred", "active", "notes"
            ];
            if (!modal) return;
            fields.forEach(function (field) {
                var input = modal.querySelector('[data-field="' + field + '"]');
                if (input) input.value = button.dataset[field] || "";
            });
            window.CatalogUI.openModal("vendorPriceEdit");
        });
    });
})();
