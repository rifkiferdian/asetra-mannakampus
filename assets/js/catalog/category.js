(function () {
    document.querySelectorAll("[data-edit-category]").forEach(function (button) {
        button.addEventListener("click", function () {
            var modal = document.getElementById("categoryEdit");
            ["id", "code", "name", "parent", "division", "description", "active"].forEach(function (field) {
                modal.querySelector('[data-field="' + field + '"]').value = button.dataset[field] || "";
            });
            Array.from(modal.querySelector('[data-field="parent"]').options).forEach(function (option) {
                option.disabled = option.value === button.dataset.id;
            });
            window.CatalogUI.openModal("categoryEdit");
        });
    });
})();
