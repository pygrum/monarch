from monarch.builder import builder_service, BuildFunction, BuildRequest, BuildResponse


def routine(r: BuildRequest) -> BuildResponse:
    b = BuildResponse(0, "", "")
    return b


svc = builder_service(BuildFunction(routine))
svc.serve_forever()