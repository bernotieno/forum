let currentPage = 1;
let isLoading = false;
let hasMoreNotifications = true;

async function fetchNotifications() {
  try {
    const response = await fetch(`/api/notifications`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": getCSRFToken(),
      },
    });

    // Clone the response so we can read the body without affecting further processing
    const responseClone = response.clone();
    const rawText = await responseClone.text();

    if (!response.ok) throw new Error("Failed to fetch notifications");

    return await response.json();
  } catch (error) {
    console.error("Error fetching notifications:", error);
    showToast("Failed to load notifications");
    throw error;
  }
}

async function markNotificationAsRead(notificationId) {
  try {
    const response = await fetch(
      `/api/notifications/read?notificationId=${notificationId}`,
      {
        method: "PUT",
      }
    );
    if (!response.ok) throw new Error("Failed to mark notification as read");
    return await response.json();
  } catch (error) {
    console.error("Error marking notification as read:", error);
    throw error;
  }
}

async function clearAllNotifications() {
  try {
    const response = await fetch("/api/notifications/clear", {
      method: "DELETE",
    });
    if (!response.ok) throw new Error("Failed to clear notifications");
    return await response.json();
  } catch (error) {
    console.error("Error clearing notifications:", error);
    throw error;
  }
}
async function initializeNotifications() {
  // Reset state
  currentPage = 1;
  isLoading = false;
  hasMoreNotifications = true;

  // Setup UI
  setupNotificationDropdown();

  // Load initial notifications
  await loadNotifications();

  // Setup event listeners only if they haven't been set up
  if (
    !document
      .querySelector(".notification-menu")
      .hasAttribute("data-initialized")
  ) {
    setupNotificationEventListeners();
    document
      .querySelector(".notification-menu")
      .setAttribute("data-initialized", "true");
  }
}

async function loadNotifications(append = false) {
  if (isLoading || (!append && !hasMoreNotifications)) return;

  isLoading = true;
  try {
    const { notifications, total, unread } = await fetchNotifications();

    // Always update badge count even if there are no notifications
    updateNotificationBadge(unread);

    if (!notifications || notifications.length === 0) {
      updateNotificationsList([], append);
      return;
    }

    updateNotificationsList(notifications, append);

    const loadMoreBtn = document.querySelector(".load-more-notifications");
    if (loadMoreBtn) {
      loadMoreBtn.style.display = hasMoreNotifications ? "block" : "none";
    }
  } catch (error) {
    console.error("Error loading notifications:", error);
  } finally {
    isLoading = false;
  }
}

function updateNotificationsList(notifications, append = false) {
  const container = document.getElementById("notifications-list");
  if (!container) return;

  if (!notifications) {
    container.innerHTML = "";
    updateNotificationBadge(0);
    return;
  }

  const notificationsHTML = notifications.map(createNotificationItem).join("");

  if (append) {
    container.insertAdjacentHTML("beforeend", notificationsHTML);
  } else {
    container.innerHTML = notificationsHTML;
  }
}

function setupNotificationEventListeners() {
  // Clear all button
  document.querySelector(".clear-all")?.addEventListener("click", async () => {
    try {
      await clearAllNotifications();
      document.getElementById("notifications-list").innerHTML = "";
      updateNotificationBadge(0);
      showToast("All notifications cleared");
    } catch (error) {
      showToast("Failed to clear notifications");
    }
  });

  // Load more button
  document
    .querySelector(".load-more-notifications")
    ?.addEventListener("click", () => {
      loadNotifications(true);
    });

  // Individual notification clicks
  document
    .getElementById("notifications-list")
    ?.addEventListener("click", handleNotificationClick);
}

async function handleNotificationClick(event) {
  const notificationItem = event.target.closest(".notification-item");
  if (!notificationItem) return;

  const notificationId = notificationItem.dataset.notificationId;

  // Handle mark as read button click
  if (event.target.closest(".mark-read-btn")) {
    try {
      await markNotificationAsRead(notificationId);
      notificationItem.classList.remove("unread");
      event.target.closest(".mark-read-btn").style.display = "none";

      // Update unread count
      const unreadCount = document.querySelectorAll(
        ".notification-item.unread"
      ).length;
      updateNotificationBadge(unreadCount);

      showToast("Notification marked as read");
    } catch (error) {
      showToast("Failed to mark notification as read", NotificationType.ERROR);
    }
    return;
  }
}

function updateNotificationBadge(count) {
  const badge = document.querySelector(".notification-badge");
  if (!badge) return;

  if (count > 0) {
    badge.textContent = count > 99 ? "99+" : count;
    badge.style.display = "block";
  } else {
    badge.textContent = "";
    badge.style.display = "none";
  }

  // Update the count in the header if it exists
  const newNotificationsCount = document.querySelector(
    ".new-notifications-count"
  );
  if (newNotificationsCount) {
    newNotificationsCount.textContent = count > 0 ? `(${count})` : "";
  }
}

function setupNotificationDropdown() {
  const notificationBtn = document.querySelector(
    ".notification-menu .icon-btn"
  );
  const dropdownMenu = document.querySelector(
    ".notification-menu .dropdown-menu"
  );
  let isHovering = false;

  if (!notificationBtn || !dropdownMenu) {
    console.error("Notification elements not found");
    return;
  }

  // Toggle dropdown on button click
  notificationBtn.addEventListener("click", (e) => {
    // Check if the click is coming from the button itself or its children
    if (e.currentTarget !== e.target && !e.target.closest(".icon-btn")) {
      return;
    }

    e.stopPropagation();
    const isVisible = dropdownMenu.classList.contains("show");

    // Close any other open dropdowns first
    document.querySelectorAll(".dropdown-menu.show").forEach((menu) => {
      if (menu !== dropdownMenu) {
        menu.classList.remove("show");
      }
    });

    dropdownMenu.classList.toggle("show");

    if (!isVisible) {
      loadNotifications();
    }
  });

  // Handle hover states
  dropdownMenu.addEventListener("mouseenter", () => {
    isHovering = true;
  });

  dropdownMenu.addEventListener("mouseleave", () => {
    isHovering = false;
    setTimeout(() => {
      if (!isHovering) {
        dropdownMenu.classList.remove("show");
      }
    }, 300);
  });

  // Close when clicking outside
  document.addEventListener("click", (e) => {
    if (
      !dropdownMenu.contains(e.target) &&
      !notificationBtn.contains(e.target)
    ) {
      dropdownMenu.classList.remove("show");
    }
  });
}
// Function to get CSRF token - add flexibility in how we find it
function getCSRFToken() {
  // Try different ways to find the CSRF token
  const csrfElement = document.querySelector('meta[name="csrf-token"]');

  if (csrfElement) {
    return csrfElement.content;
  }

  // If no token found, return null or an empty string
  return "";
}
function getNotificationTypeIcon(type) {
  const icons = {
    like: '<i class="fas fa-heart"></i>',
    comment: '<i class="fas fa-comment"></i>',
    follow: '<i class="fas fa-user-plus"></i>',
    mention: '<i class="fas fa-at"></i>',
    default: '<i class="fas fa-bell"></i>',
  };
  return icons[type] || icons.default;
}
function createNotificationItem(notification) {
  return `
        <div class="notification-item ${notification.is_read ? "" : "unread"}" 
             data-notification-id="${notification.id}">
            <div class="notification-icon">
                ${getNotificationTypeIcon(notification.type)}
            </div>
            <div class="notification-content">
                <p><strong>${escapeHTML(
                  notification.actor.nickname
                )}</strong> ${escapeHTML(notification.message)}</p>
                <span class="notification-time">${formatTimeAgo(
                  notification.created_at
                )}</span>
            </div>
            <div class="notification-actions">
                <button class="mark-read-btn" title="Mark as read" ${
                  notification.is_read ? 'style="display: none;"' : ""
                }>
                    <i class="fas fa-check"></i>
                </button>
            </div>
        </div>
    `;
}
// XSS Prevention utilities
function escapeHTML(str) {
  if (!str) return "";
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}
function formatTimeAgo(timestamp) {
  const date = new Date(timestamp);
  const now = new Date();
  const diff = Math.floor((now - date) / 1000);

  if (diff < 60) return "Just now";
  if (diff < 3600) return `${Math.floor(diff / 60)}min`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}hrs`;
  if (diff < 31536000) return `${Math.floor(diff / 86400)}days`;
  return `${Math.floor(diff / 31536000)}yrs`;
}
initializeNotifications();
