console.log("hello world");
console.log("loading jquery...");
$(function () {
    console.log("✅ jQuery loaded successfully. Version:", jQuery.fn.jquery);
    if (window.Chart) {
        console.log("✅ Chart.js loaded successfully. Version:", Chart.version);
    } else {
        console.log("❌ Chart.js is not loaded.");
    }
    if (window.moment) {
        console.log("✅ Moment.js loaded successfully. Version:", moment.version);
    } else {
        console.log("❌ Moment.js is not loaded");
    }




});