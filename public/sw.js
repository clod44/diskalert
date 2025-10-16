self.addEventListener('push', (event) => {
    let payload = {
        title: 'Disk Alert',
        body: 'Disk usage alert received, but payload was empty.',
        icon: '/favicon.ico',
        badge: '/favicon.ico',
    };

    if (event.data) {
        try {
            const data = event.data.json();
            payload.title = data.title || payload.title;
            payload.body = data.message || payload.body;
        } catch (e) {
            payload.body = event.data.text();
        }
    }

    event.waitUntil(
        self.registration.showNotification(payload.title, {
            body: payload.body,
            icon: payload.icon,
            badge: payload.badge,
            vibrate: [200, 100, 200]
        })
    );
});

self.addEventListener('notificationclick', (event) => {
    event.notification.close();
    // Action: Open the main dashboard when the user clicks the notification
    event.waitUntil(
        clients.openWindow('/')
    );
});

self.addEventListener('install', (event) => {
    // Force the new Service Worker to activate immediately
    event.waitUntil(self.skipWaiting());
});

self.addEventListener('activate', (event) => {
    // Claim control of all pages under its scope
    event.waitUntil(self.clients.claim());
});