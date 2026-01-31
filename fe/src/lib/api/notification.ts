import apiClient from "../api-client";




export async function getNotifications(limit = 20, offset = 0) {
    try {
        const token = localStorage.getItem("access_token");

        if(!token) {
            throw new Error("No access token found");
        }

        const response = await apiClient.get<{ data: Notification[] }>(`/api/v1/notifications`, {
            headers: {
                Authorization: `Bearer ${token}`,
            },
            params: { limit, offset },
        });
        return response.data.data;

    } catch (error) {
        console.error("Error fetching notifications:", error);
        throw error;
    }
}