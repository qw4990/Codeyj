package main

func MimeType(path string) string {
    l := len(path)
    if l > 4 && path[l-4:] == "html" {
        return "text/html"
    } else if l > 2 && path[l-2:] == "js" {
        return "application/x-javascript"
    } else if l > 3 && path[l-3:] == "css" {
        return "text/css"
    } else if l > 3 && path[l-3:] == "png" {
        return "image/png"
    } else if l > 4 && path[l-4:] == "jpeg" {
        return "image/jpeg"
    } else {
        return "text/plain"
    }
}
