(function () {
    function openModal(id) {
        var modal = document.getElementById(id);
        if (!modal) return;
        modal.classList.remove("hidden");
        modal.classList.add("flex");
        document.body.classList.add("overflow-hidden");
    }

    function closeModals() {
        document.querySelectorAll(".catalog-backdrop").forEach(function (modal) {
            modal.classList.add("hidden");
            modal.classList.remove("flex");
        });
        document.body.classList.remove("overflow-hidden");
    }

    document.querySelectorAll("[data-open]").forEach(function (button) {
        button.addEventListener("click", function () {
            openModal(button.dataset.open);
        });
    });

    document.querySelectorAll("[data-close]").forEach(function (button) {
        button.addEventListener("click", closeModals);
    });

    document.querySelectorAll("[data-confirm]").forEach(function (form) {
        form.addEventListener("submit", function (event) {
            if (!window.confirm(form.dataset.confirm)) event.preventDefault();
        });
    });

    document.addEventListener("keydown", function (event) {
        if (event.key === "Escape") closeModals();
    });

    window.CatalogUI = {
        openModal: openModal,
        closeModals: closeModals
    };
})();
