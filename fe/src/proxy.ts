import { NextRequest } from "next/server";

export function proxy (request: NextRequest) {

    

    const { pathname } = request.nextUrl;

    // Guard 1 when user is having access token but trying to access auth routes
    const accessToken = request.cookies.get("access_token")?.value;
    if (accessToken) {
        const authRoutes = ["/sign-in", "/sign-up", "/forgot-password"];    
        if (authRoutes.includes(pathname)) {
            const url = request.nextUrl.clone();
            url.pathname = "/dashboard";
            return Response.redirect(url);
        }
      return;   
    }

    // Guard 2 when user is not having access token but trying to access protected routes
    const protectedRoutes = ["/dashboard", "/profile", "/settings"];
    if (!accessToken || accessToken === "") {
        if (protectedRoutes.includes(pathname)) {
            const url = request.nextUrl.clone();
            url.pathname = "/sign-in";
            return Response.redirect(url);
        }   
    } return;


    // if not have any token and trying to access auth routes, allow

    


}