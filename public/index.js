async function fetchStats() {
    try {
        const response = await fetch('/api/stats');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        console.log(data);
        const formattedJson = JSON.stringify(data, null, 2);
        document.getElementById("stats").textContent = formattedJson;
    } catch (error) {
        console.error("Error fetching metrics:", error);
    }
}
fetchStats();
setInterval(fetchStats, 5000);

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