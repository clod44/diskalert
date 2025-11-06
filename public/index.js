console.log("hello world");
console.log("loading jquery...");
$(async () => {
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

    $('#subscribe-btn').on('click', handleSubscribe);
    $('#unsubscribe-btn').on('click', handleUnsubscribe);
    $('#subscriptions-btn').on('click', handleShowSubscriptions)


    $("#subscribe-btn").on('click', (event) => {
        event.preventDefault();
        handleSubscribe();
    });
    $("#unsubscribe-btn").on('click', (event) => {
        event.preventDefault();
        handleUnsubscribe();
    });
    $("#subscriptions-btn").on('click', (event) => {
        event.preventDefault();
        handleShowSubscriptions();
    });

    $("#test-notification-all").on('click', (event) => {
        event.preventDefault();
        handleTestNotification();
    });

    $("#test-notification").on('click', (event) => {
        event.preventDefault();
        const endpoint = $("#test-notification-endpoint").value;
        handleTestNotification(endpoint);
    });

    await showExistingSubscriptionDetails()
    await showDiskGraphs();
});

const showDiskGraphs = async () => {
    console.log("attempting toshow disk graphs..", window.HOMEDATA)
    window.HOMEDATA.DiskStatus.records.forEach(disk => {
        console.log(disk)
        showDiskGraph(disk.uuid);
    });
}

const showDiskGraph = async (uuid) => {
    await showDiskHistory(uuid)
    await showDiskForecast(uuid)
}
const showDiskHistory = async (uuid) => {
    const data = await getDiskHistory(uuid);
    const diskPath = data[0].DiskPath;
    const usedSizeData = data.map(record => ({
        x: record.Timestamp * 1000,
        y: record.UsedSize
    }));
    const totalSizeData = data.map(record => ({
        x: record.Timestamp * 1000,
        y: record.TotalSize
    }));
    createChart(diskPath, {
        label: 'Used Size (Actual)',
        data: usedSizeData,
        borderColor: 'rgba(255, 0, 0, 1)',
        backgroundColor: 'rgba(200, 0, 0, 0.2)',
        borderWidth: 4,
        fill: false,
        pointRadius: 2,
        hitRadius: 5,
    }, uuid);
    createChart(diskPath, {
        label: 'Total Size',
        data: totalSizeData,
        borderColor: 'rgba(30, 255, 0, 1)',
        backgroundColor: 'rgba(23, 200, 0, 0.2)',
        borderWidth: 2,
        fill: false,
        pointRadius: 2,
        hitRadius: 5,
    }, uuid);
}
const showDiskForecast = async (uuid) => {
    const data = await getDiskForecast(uuid);
    data.forecasts.forEach((forecast, index, arr) => {
        //for each different forecast algorithm result:
        const dataPoints = [];
        const pointColors = [];
        forecast.data.forEach(point => {
            dataPoints.push({
                x: point.x * 1000,
                y: point.y
            });
            pointColors.push(point.color || 'rgba(0, 89, 255, 1)');
        });
        createChart(null, {
            label: forecast.algorithmName,
            borderWidth: 3,
            data: dataPoints,
            pointBackgroundColor: pointColors,
            borderColor: pointColors[0],
            backgroundColor: 'rgba(0, 140, 255, 0.2)',
            fill: false,
            hitRadius: 5,
            pointRadius: 2
        }, uuid);
    })
}
const getDiskHistory = async (uuid) => {
    try {
        const response = await fetch(`/api/history?uuid=${uuid}&limit=30`, {
            method: 'GET'
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        console.log(data);
        return data;
    } catch (error) {
        console.error("Error fetching disk records history:", error);
        return null
    }
}
const getDiskForecast = async (uuid) => {
    try {
        //all forecasts data comes with simple point datas {x,y} x as timestamp. y as usedsize
        const response = await fetch(`/api/forecast?uuid=${uuid}&limit=30`, {
            method: 'GET'
        });
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const forecasts = await response.json();
        if (!forecasts || forecasts.length === 0) {
            console.log(`No forecasts data available for ${uuid}.`);
            return;
        }
        console.log(forecasts)
        return forecasts;
    } catch (error) {
        console.error("Error fetching disk records history:", error);
        return null
    }
}



const chartInstances = {};

const destroyAllCharts = () => {
    for (const uuid in chartInstances) {
        if (Object.hasOwnProperty.call(chartInstances, uuid)) {
            const chart = chartInstances[uuid];
            if (chart && typeof chart.destroy === 'function') {
                chart.destroy();
                delete chartInstances[uuid];
            }
        }
    }
    console.log("All previous Chart.js instances have been destroyed.");
}

const formatBytes = (bytes) => {
    if (!Number.isFinite(bytes) || bytes < 0) {
        return bytes;
    }
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}
const sizeTooltipFormatter = (tooltipItem) => {
    const value = tooltipItem.parsed.y;
    if (isNaN(value)) {
        return `${tooltipItem.dataset.label}: Data Not Available`;
    }
    return `${tooltipItem.dataset.label}: ${formatBytes(value)}`;
};

/**
 * Creates or updates a Chart.js instance in a specified container.
 * It handles adding or updating datasets based on their 'label'.
 *
 * @param {string} title The main title for the chart.
 * @param {object} newDataset The dataset object to add or update (must contain a 'label' property).
 * @param {string} id The ID of the HTML container element (e.g., a div) where the chart should live.
 */
const createChart = (title, newDataset, id) => {
    const statContainer = document.getElementById(id);
    if (!statContainer) {
        console.error(`statContainer element not found for ID: ${id}`);
        return;
    }
    let container = document.getElementById("chart-wrapper-" + id);
    if (!container) {
        container = document.createElement('div');
        container.id = "chart-wrapper-" + id;
        statContainer.appendChild(container);
    }
    let canvas = container.querySelector('canvas');
    let existingChart = canvas ? Chart.getChart(canvas) : null;

    // --- CHART CREATION LOGIC ---
    if (!existingChart) {
        if (!canvas) {
            canvas = document.createElement('canvas');
            canvas.id = `chart-canvas${id}`;
            canvas.style.height = '256px';
            container.appendChild(canvas);
        }
        const config = {
            type: 'line',
            data: {
                datasets: [newDataset]
            },
            options: {
                interaction: {
                    mode: 'nearest',
                    intersect: true
                },
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Disk Size (Bytes)'
                        },
                    },
                    x: {

                        type: 'linear',
                        ticks: {
                            callback: function (value, index, ticks) {
                                const date = new Date(value);
                                return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
                            }
                        },
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                },
                plugins: {
                    title: {
                        display: true,
                        text: title
                    }
                }
            }
        };
        new Chart(canvas, config);
        return;
    }

    // --- CHART UPDATE LOGIC ---
    let found = false;
    const datasets = existingChart.data.datasets;
    for (let i = 0; i < datasets.length; i++) {
        if (datasets[i].label === newDataset.label) {
            datasets[i] = newDataset;
            found = true;
            break;
        }
    }
    if (!found) {
        existingChart.data.datasets.push(newDataset);
    }
    if (title) existingChart.options.plugins.title.text = title;
    existingChart.update();
}




/**
 * Converts a URL-safe Base64 string (like a VAPID key) into a Uint8Array.
 * @param {string} base64String - The URL-safe Base64 string.
 * @returns {Uint8Array}
 */
const urlBase64ToUint8Array = (base64String) => {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
    const rawData = atob(base64);
    return Uint8Array.from(rawData, char => char.charCodeAt(0));
};

/**
 * Handles the push notification subscription process end-to-end.
 */
const handleSubscribe = async () => {
    const $subBtn = $('#subscribe-btn');
    $subBtn.prop('disabled', true);
    const $unsubBtn = $('#unsubscribe-btn');
    if (!navigator.serviceWorker || !window.PushManager) {
        console.warn('Push messaging is not supported by this browser.');
        console.error("Browser does not support push web notifications (original alert suppressed).");
        return;
    }
    $subBtn.prop('disabled', true);
    try {
        const reg = await navigator.serviceWorker.register('/public/sw.js', { scope: '/' });
        const vapidResp = await fetch('/api/vapid-key');
        if (!vapidResp.ok) throw new Error('Failed to fetch VAPID key.');
        const key = await vapidResp.text();
        const convertedKey = urlBase64ToUint8Array(key);
        const sub = await reg.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: convertedKey,
        });
        const sendResp = await fetch('/api/subscribe', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(sub),
        });
        if (!sendResp.ok) throw new Error('Server failed to save subscription.');
    } catch (error) {
        console.error('Subscription failed:', error);
    } finally {
        $subBtn.prop('disabled', false);
        $unsubBtn.prop('disabled', false);
    }
};

/**
 * Handles the push notification unsubscription process.
 */
const handleUnsubscribe = async () => {
    const $subBtn = $('#subscribe-btn');
    const $unsubBtn = $('#unsubscribe-btn');
    if (!navigator.serviceWorker || !window.PushManager) return;
    $unsubBtn.prop('disabled', true);
    try {
        const reg = await navigator.serviceWorker.getRegistration('/');
        if (!reg) return;
        const sub = await reg.pushManager.getSubscription();
        if (!sub) return;
        await sub.unsubscribe();
        const sendResp = await fetch('/api/unsubscribe', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(sub),
        });
        if (!sendResp.ok) throw new Error('Server deletion failed.');
    } catch (error) {
        console.error('Unsubscribe process error:', error);
    } finally {
        $subBtn.prop('disabled', false);
        $unsubBtn.prop('disabled', false);
    }
};


/**
 * Handles the process of displaying all the subscriptions stored on the server.
 * @throws {Error} If an error occurs during the process of fetching the subscriptions from the server.
 */
const handleShowSubscriptions = async () => {
    const $subBtn = $("#subscriptions-btn");
    $subBtn.prop('disabled', true);
    const $subscriptions = $("#subscriptions");
    try {
        const response = await fetch('/api/subscriptions');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        console.log(data);
        const formattedJson = JSON.stringify(data, null, 2);
        $subscriptions.text(formattedJson);
    } catch (error) {
        console.error('cant get subscriptions. error:', error);
    } finally {
        $subBtn.prop('disabled', false);
    }
}

/**
 * Fetches the existing push subscription details using modern async/await syntax.
 * @returns {Promise<object|null>} The subscription details as JSON or null if none exists or Service Worker is not supported.
 */
const getExistingSubscriptionDetails = async () => {
    if (!('serviceWorker' in navigator)) {
        return null;
    }
    try {
        const registration = await navigator.serviceWorker.ready;
        const subscription = await registration.pushManager.getSubscription();
        if (subscription) {
            return subscription.toJSON();
        } else {
            return null;
        }
    } catch (error) {
        console.error("Error fetching subscription details:", error);
        throw error;
    }
}

const showExistingSubscriptionDetails = async () => {
    //show the current subscription data in the client
    const $subscriptionDataElement = $("#subscription-data");
    try {
        const data = await getExistingSubscriptionDetails();
        if (data) {
            const formattedJson = JSON.stringify(data, null, 2);
            $subscriptionDataElement.text(formattedJson);
            console.log("Subscription details displayed.");
        } else {
            $subscriptionDataElement.text("No active push subscription found.");
            console.log("No active push subscription found.");
        }
    } catch (error) {
        $subscriptionDataElement.text("ERROR: Failed to fetch subscription details.");
        console.error("FATAL ERROR in subscription logic:", error);
    }
}

const handleTestNotification = async (endpoint = null) => {
    const options = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            target_endpoint: endpoint,
        })
    };
    try {
        const response = await fetch('/api/test-notification', options);
        if (!response.ok) {
            const errorBody = await response.text();
            throw new Error(`HTTP error! Status: ${response.status}. Body: ${errorBody}`);
        }
        console.log('Notification test successfully initiated.');
    } catch (error) {
        console.error('Error initiating subscription test:', error);
    }
}
