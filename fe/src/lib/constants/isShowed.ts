export const shouldShowSidebar = (pathname: string): boolean => {
  const hideSidebarPatterns = [
    /^\/dashboard\/pod\/create$/,
    /^\/dashboard\/pods\/[^/]+$/,
  ];
    return !hideSidebarPatterns.some(pattern => pattern.test(pathname));
};

export const shouldShowRightSidebar = (pathname: string): boolean => {
  const hideRightSidebarPatterns = [
    /^\/dashboard\/my-pods\/[^/]+$/,
    /^\/dashboard\/pods\/[^/]+$/,
    /^\/dashboard\/pod\/create$/,
    /^\/dashboard\/my-pods$/,
  ];
    return !hideRightSidebarPatterns.some(pattern => pattern.test(pathname));
};

