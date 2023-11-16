from http.server import BaseHTTPRequestHandler, HTTPServer
import json

STATUS_SUCCESS = 0
STATUS_ERROR = 1


class BuildRequest:
    def __init__(self, params: dict):
        """
        An object representing a build request received from the main C2 server.
        :param params: a dictionary of build parameters.
        """
        self.params = params


class BuildResponse:
    def __init__(self, status: int, error: str, build: bytes):
        """
        An object representing a response to a build request received from the main C2 server.
        A build would typically end with a status (STATUS_SUCCESS | STATUS_ERROR), an error message (if applicable),
        and the binary or file produced by the build.
        :param status: build status, either STATUS_SUCCESS or STATUS_ERROR
        :param error: build error message, if any
        :param build: the build file. If multiple files are produced, then return an archive.
        """
        self.status = status
        self.error = error
        self.build = build


class BuildFunction:
    def __init__(self, routine):
        """
        BuildFunction is initialized with a routine that takes a BuildRequest object, performs a build, and returns
        a BuildResponse object.
        :param routine: a build routine (function) that receives build parameters in a BuildRequest object and returns
        a BuildResponse, which includes build status, the actual build, and any potential error messages
        """
        self.routine = routine

    def __call__(self, request: BuildRequest) -> BuildResponse:
        return self.routine(BuildRequest)


class MonarchBuilder(BaseHTTPRequestHandler):

    build: BuildFunction

    def register_build(self, function: BuildFunction):
        self.build = function

    def do_POST(self):
        content_length = int(self.headers.get('Content-Length'))

        post_data = self.rfile.read(content_length)
        data = json.loads(post_data)

        if self.path.startswith("/build"):
            # data is a dict<string, string> as defined in protobuf
            request = BuildRequest(params=data)
            if not hasattr(self, "build"):
                response_json = {
                    "status": STATUS_ERROR,
                    "error": "build routine has not been registered."
                }
            else:
                response_object = self.build(request)
                response_json = {
                    "status": response_object.status,
                    "error": response_object.error,
                    "build": response_object.build
                }
            response = json.dumps(response_json)
            self.send_response(200)
            self.send_header("content-type", "application/json")
            self.end_headers()
            self.wfile.write(bytes(response, "utf-8"))
        else:
            self.send_response(404)


def builder_service() -> HTTPServer:
    """
    :return: A HTTPServer class using the monarch translator class as a request handler
    """
    service_address = ("localhost", 20000)
    return HTTPServer(service_address, MonarchBuilder)
