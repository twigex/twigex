// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// /src/services/authService.js
import axios from "./api";

const authService = {
    // Function to get user details
    logIn(loginID, password, token) {
        return axios.post(
            "/auth/login",
            { login_id: loginID, password: password, token: token },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    logOut() {
        return axios.post("/auth/logout");
    },

    resetPassword(email) {
        return axios.post("/password/reset", { email: email });
    },

    updatePassword(data) {
        return axios.post(`/password/reset/${data.token}`, {
            password: data.password,
            confirmPassword: data.confirmPassword,
        });
    },

    passwordPolicy() {
        return axios.get("/password/policy");
    },
};

export default authService;
