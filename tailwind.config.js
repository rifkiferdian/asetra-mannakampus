module.exports = {
  content: ["./templates/**/*.html"],
  theme: {
    extend: {
      fontFamily: {
        display: ["Plus Jakarta Sans", "ui-sans-serif", "system-ui"],
        body: ["Manrope", "ui-sans-serif", "system-ui"],
        "display-auth": ["Space Grotesk", "ui-sans-serif", "system-ui"],
      },
      colors: {
        brand: {
          50: "#fff7ed",
          100: "#ffedd5",
          300: "#fdba74",
          500: "#f26419",
          600: "#dc5a15",
          700: "#9a3412",
        },
        "brand-auth": {
          50: "#fff7ed",
          100: "#ffedd5",
          200: "#fed7aa",
          300: "#fdba74",
          400: "#fb923c",
          500: "#f97316",
          600: "#ea580c",
          700: "#c2410c",
          800: "#9a3412",
          900: "#7c2d12",
        },
      },
      boxShadow: {
        card: "0 12px 26px rgba(16, 24, 40, 0.08)",
        glow: "0 30px 80px -40px rgba(242, 100, 25, 0.45)",
      },
    },
  },
  plugins: [],
};
