package handlers

import (
"encoding/json"
"net/http"

"github.com/Traderong/omnivibe-api/auth"
)

func Me(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

userID, ok := auth.UserIDFromContext(r.Context())
if !ok {
http.Error(w, "authentication required", http.StatusUnauthorized)
return
}

user, err := auth.GetUserByID(r.Context(), userID)
if err != nil {
http.Error(w, "user not found", http.StatusNotFound)
return
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)

_ = json.NewEncoder(w).Encode(user)
}
