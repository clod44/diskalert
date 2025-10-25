document.addEventListener('DOMContentLoaded', () => {
    function fetchExistingSubscriptionDetails() {
        return new Promise((resolve, reject) => {
            if (!('serviceWorker' in navigator)) {
                resolve(null);
                return;
            }
            navigator.serviceWorker.ready.then(registration => {
                registration.pushManager.getSubscription().then(subscription => {
                    if (subscription) {
                        resolve(subscription.toJSON());
                    } else {
                        resolve(null);
                    }
                }).catch(error => {
                    reject(error);
                });
            }).catch(error => {
                reject(error);
            });
        });
    }
    const existingSubscription = fetchExistingSubscriptionDetails();
    existingSubscription.then(data => {
        const formattedJson = JSON.stringify(data, null, 2);
        document.getElementById("subscription-data").textContent = formattedJson;
    });

    async function fetchStats() {
        try {
            const response = await fetch("/api/stats", {
                method: 'GET'
            });
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            console.log(data);
            const statsDiv = document.getElementById("stats");
            statsDiv.innerHTML = '' //delete all children
            destroyAllCharts();
            //create children again
            data.records.forEach(record => {
                const div = document.createElement("div");
                div.id = record.uuid;
                div.style.display = "flex";
                div.style.gap = "20px";
                div.style.overflow = "hidden";
                const pre = document.createElement("pre");
                pre.style.flexShrink = "0";
                pre.style.width = "400px";
                pre.textContent = JSON.stringify(record, null, 2);
                div.appendChild(pre);
                statsDiv.appendChild(div);
                showDiskHistory(record.uuid);
            });
        } catch (error) {
            console.error("Error fetching metrics:", error);
        }
    }
    fetchStats();
    var fetchStatusDuration = 60000;
    setInterval(fetchStats, fetchStatusDuration);
    const chartInstances = {};

    function destroyAllCharts() {
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

    function formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
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
    function createChart(title, newDataset, id) {
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
                canvas.id = `${id}-canvas`;
                canvas.style.height = '256px';
                container.appendChild(canvas);
            }
            const config = {
                type: 'line',
                data: {
                    datasets: [newDataset]
                },
                options: {
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
                            type: 'time',
                            time: {
                                unit: 'hour',
                                tooltipFormat: 'MMM DD, HH:mm:ss',
                                displayFormats: {
                                    hour: 'HH:mm',
                                    day: 'MMM DD'
                                }
                            },
                            title: {
                                display: true,
                                text: 'Time'
                            }
                        }
                    },
                    plugins: {
                        legend: {
                            display: true,
                            position: 'bottom',
                        },
                        tooltip: {
                            callbacks: {
                                label: sizeTooltipFormatter
                            }
                        },
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
    function _createDiskChart(uuid, historyRecords, diskPath) {
        const historicalRecords = historyRecords.filter(record => !record.Forecast);
        const forecastRecords = historyRecords.filter(record => record.Forecast);
        const usedHistoricalData = historicalRecords.map(record => ({
            x: record.Timestamp * 1000,
            y: record.UsedSize
        }));
        let usedForecastData = [];
        if (historicalRecords.length > 0) {
            //so the forecast graph line and the actual graph line connects
            const newestHistoryPoint = usedHistoricalData[0];
            usedForecastData.push(newestHistoryPoint);

            forecastRecords.forEach(record => {
                usedForecastData.push({
                    x: record.Timestamp * 1000,
                    y: record.UsedSize
                });
            });
        }
        const div = document.getElementById(uuid);
        const totalData = historyRecords.map(record => ({
            x: record.Timestamp * 1000,
            y: record.TotalSize
        }));

        let chartWrapper = div.querySelector(`#chart-wrapper-${uuid}`);
        if (!chartWrapper) {
            chartWrapper = document.createElement('div');
            chartWrapper.id = `chart-wrapper-${uuid}`;
            let placeholder = div.querySelector('div:last-child');
            if (placeholder && placeholder.textContent.includes('history')) {
                placeholder.replaceWith(chartWrapper);
            } else {
                div.appendChild(chartWrapper);
            }
        }
        if (chartInstances[uuid]) {
            chartInstances[uuid].destroy();
            delete chartInstances[uuid];
        }
        chartWrapper.innerHTML = '';
        let canvas = document.createElement('canvas');
        canvas.id = `chart-${uuid}`;
        canvas.style.width = '100%';
        canvas.style.height = '256px';
        chartWrapper.appendChild(canvas);
        const ctx = canvas.getContext('2d');

        chartInstances[uuid] = new Chart(ctx, {
            type: 'line',
            data: {
                datasets: [
                    {
                        label: 'Used Size (Actual)',
                        data: usedHistoricalData,
                        borderColor: 'rgba(255, 0, 0, 1)',
                        backgroundColor: 'rgba(200, 0, 0, 0.2)',
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 2,
                    },
                    {
                        label: 'Used Size (Forecast)',
                        data: usedForecastData,
                        borderColor: 'rgba(255, 165, 0, 1)',
                        backgroundColor: 'transparent',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        fill: false,
                        pointRadius: 1,
                    },
                    {
                        label: 'Total Size',
                        data: totalData,
                        borderColor: 'rgba(0, 255, 0, 1)',
                        backgroundColor: 'rgba(0, 200, 0, 0.2)',
                        borderWidth: 2,
                        fill: false,
                        pointRadius: 0,
                    }
                ]
            },
            options: {
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
                        type: 'time',
                        time: {
                            unit: 'hour',
                            tooltipFormat: 'MMM DD, HH:mm:ss',
                            displayFormats: {
                                hour: 'HH:mm',
                                day: 'MMM DD'
                            }
                        },
                        title: {
                            display: true,
                            text: 'Time'
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: true,
                        position: 'bottom',
                    },
                    tooltip: {
                        callbacks: {
                            label: sizeTooltipFormatter
                        }
                    },
                    title: {
                        display: true,
                        text: `${diskPath} Size Trends`
                    }
                }
            }
        });
    }
    async function showDiskHistory(uuid) {
        try {
            const response = await fetch(`/api/history?uuid=${uuid}&limit=30`, {
                method: 'GET'
            });
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const historyRecords = await response.json();
            if (!historyRecords || historyRecords.length === 0) {
                console.log(`No history data available for ${uuid}.`);
                const div = document.getElementById(uuid);
                if (div) {
                    const chartWrapper = div.querySelector(`#chart-wrapper-${uuid}`);
                    if (chartWrapper) {
                        chartWrapper.innerHTML = `No historical data available for ${uuid}.`;
                    }
                }
                return;
            }
            console.log(historyRecords)
            const diskPath = historyRecords[0].DiskPath;
            const historyData = historyRecords.map(record => ({
                x: record.Timestamp * 1000,
                y: record.UsedSize
            }));
            const totalData = historyRecords.map(record => ({
                x: record.Timestamp * 1000,
                y: record.TotalSize
            }));
            createChart(diskPath, {
                label: 'Used Size (Actual)',
                data: historyData,
                borderColor: 'rgba(255, 0, 0, 1)',
                backgroundColor: 'rgba(200, 0, 0, 0.2)',
                borderWidth: 2,
                fill: false,
                pointRadius: 2,
            }, uuid);
            createChart(diskPath, {
                label: 'Total Size',
                data: totalData,
                borderColor: 'rgba(30, 255, 0, 1)',
                backgroundColor: 'rgba(23, 200, 0, 0.2)',
                borderWidth: 2,
                fill: false,
                pointRadius: 2,
            }, uuid);
            showDiskForecast(uuid);
        } catch (error) {
            console.error("Error fetching disk records history:", error);
            const div = document.getElementById(uuid);
            if (div) {
                const chartWrapper = div.querySelector(`#chart-wrapper-${uuid}`);
                if (chartWrapper) {
                    chartWrapper.innerHTML = `Error loading history: ${error.message}`;
                } else {
                    let placeholder = div.querySelector('div:last-child');
                    if (placeholder) {
                        placeholder.textContent = `Error loading history: ${error.message}`;
                    }
                }
            }
        }
    }
    async function showDiskForecast(uuid) {
        try {
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
            forecasts.forecasts.forEach((forecast, index, arr) => {
                const data = forecast.data.map(record => ({
                    x: record.Timestamp * 1000,
                    y: record.UsedSize
                }));
                createChart(null, {
                    label: forecast.algorithmName,
                    data: data,
                    borderColor: 'rgba(0, 89, 255, 1)',
                    backgroundColor: 'rgba(0, 140, 255, 0.2)',
                    borderWidth: 2,
                    fill: false,
                    pointRadius: 2,
                }, uuid);
            })
        } catch (error) {
            console.error("Error fetching disk records history:", error);
            const div = document.getElementById(uuid);
            if (div) {
                const chartWrapper = div.querySelector(`#chart-wrapper-${uuid}`);
                if (chartWrapper) {
                    chartWrapper.innerHTML = `Error loading history: ${error.message}`;
                } else {
                    let placeholder = div.querySelector('div:last-child');
                    if (placeholder) {
                        placeholder.textContent = `Error loading history: ${error.message}`;
                    }
                }
            }
        }
    }
    function urlBase64ToUint8Array(base64String) {
        const padding = '='.repeat((4 - base64String.length % 4) % 4);
        const base64 = (base64String + padding)
            .replace(/-/g, '+')
            .replace(/_/g, '/');
        const rawData = window.atob(base64);
        const outputArray = new Uint8Array(rawData.length);
        for (let i = 0; i < rawData.length; ++i) {
            outputArray[i] = rawData.charCodeAt(i);
        }
        return outputArray;
    }

    async function handleSubscribe() {
        const subscribeBtn = document.getElementById("subscribe-btn");
        if (!('serviceWorker' in navigator) || !('PushManager' in window)) {
            console.warn('Push messaging is not supported by this browser.');
            alert("this browser does not support push web notifications")
            return;
        }
        try {
            const registration = await navigator.serviceWorker.register('/sw.js');
            console.log('Service Worker registered successfully.');
            const vapidResponse = await fetch('/api/vapid-key');
            if (!vapidResponse.ok) {
                throw new Error('Failed to fetch VAPID key from server.');
            }
            const applicationServerKey = await vapidResponse.text();
            const convertedVapidKey = urlBase64ToUint8Array(applicationServerKey);
            subscription = await registration.pushManager.subscribe({
                userVisibleOnly: true,
                applicationServerKey: convertedVapidKey,
            });
            console.log('Push subscription successful. Sending to server.');
            const sendResponse = await fetch('/api/subscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(subscription),
            });
            if (sendResponse.ok) {
                console.log('Subscription successfully saved on the server.');
            } else {
                throw new Error('Server failed to save subscription.');
            }
        } catch (error) {
            console.error('Subscription failed:', error);
        } finally {
            subscribeBtn.disabled = false;
            unsubscribeBtn.disabled = false;
        }
    }

    async function handleUnsubscribe() {
        const unsubscribeBtn = document.getElementById("unsubscribe-btn");
        if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;
        try {
            const registration = await navigator.serviceWorker.getRegistration('/sw.js');
            if (!registration) return;
            const subscription = await registration.pushManager.getSubscription();
            if (!subscription) return;
            await subscription.unsubscribe();
            const sendResponse = await fetch('/api/unsubscribe', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(subscription),
            });
            if (!sendResponse.ok) throw new Error('Server deletion failed.');
        } catch (error) {
            console.error('Unsubscribe process error:', error);
        } finally {
            unsubscribeBtn.disabled = false;
            subscribeBtn.disabled = false;
        }
    }

    async function handleSubscriptions() {
        const subscriptionsBtn = document.getElementById("subscriptions-btn");
        try {
            const response = await fetch('/api/subscriptions');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            console.log(data);
            const formattedJson = JSON.stringify(data, null, 2);
            document.getElementById("subscriptions").textContent = formattedJson;
        } catch (error) {
            console.error('cant get subscriptions. error:', error);
        } finally {
            subscriptionsBtn.disabled = false;
        }
    }

    async function handleTestNotification(endpoint = null) {
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
    const subscribeBtn = document.getElementById("subscribe-btn");
    subscribeBtn.addEventListener('click', (event) => {
        event.preventDefault();
        subscribeBtn.disabled = true;
        handleSubscribe();
    });
    const unsubscribeBtn = document.getElementById("unsubscribe-btn");
    unsubscribeBtn.addEventListener('click', (event) => {
        event.preventDefault();
        unsubscribeBtn.disabled = true;
        handleUnsubscribe();
    });
    const subscriptionsBtn = document.getElementById("subscriptions-btn");
    subscriptionsBtn.addEventListener('click', (event) => {
        event.preventDefault();
        subscriptionsBtn.disabled = true;
        handleSubscriptions();
    });

    const testNotificationAllBtn = document.getElementById("test-notification-all");
    testNotificationAllBtn.addEventListener('click', (event) => {
        event.preventDefault();
        handleTestNotification();
    });

    const testNotificationBtn = document.getElementById("test-notification");
    testNotificationBtn.addEventListener('click', (event) => {
        event.preventDefault();
        const endpoint = document.getElementById("test-notification-endpoint").value;
        handleTestNotification(endpoint);
    });

})
