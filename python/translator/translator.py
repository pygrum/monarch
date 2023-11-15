from http.server import BaseHTTPRequestHandler, HTTPServer
import json


class TranslateToRequest:
    def __init__(self, agent_id: str, request_id: str, opcode: int, args: list[bytes]):
        """
        An object representing a request received from the C2 server to translate a message to be sent to an agent.
        :param agent_id: The agent ID
        :param request_id: The request ID
        :param opcode: The opcode for the command the operator wants to run on the agent
        :param args: Arguments to the command
        """
        self.agent_id = agent_id
        self.request_id = request_id
        self.opcode = opcode
        self.args = args


class TranslateToResponse:
    def __init__(self, success: bool, error_msg: str, message: bytes):
        """
        An object representing the reply to a translation request from the C2
        :param success: Indicates whether the translation succeeded without errors
        :param error_msg: If the translation was unsuccessful, the error message is populated.
        :param message: The actual translation
        """
        self.success = success
        self.error_msg = error_msg
        self.message = message


class TranslateFromRequest:
    def __init__(self, data: bytes):
        """
        An object representing a translation request for received agent messages
        :param data: the data received in agent response
        """
        self.data = data


class Response:
    def __init__(self, data: bytes, destination: int):
        self.data = data
        self.destination = destination


class TranslateFromResponse:
    def __init__(self, agent_id: str, request_id: str, status: int, responses: list[Response]):
        """
        An object representing a reply to a translation request from the C2 for decoding a message received from
        an agent.
        :param agent_id: the agent ID
        :param request_id: the request ID
        :param status:
        :param responses: the data received from the agent
        """
        self.agent_id = agent_id
        self.request_id = request_id
        self.status = status
        self.responses = responses


class TranslateToFunction:
    def __init__(self, routine):
        """
        :param routine: A translation routine that, given a TranslateToRequest object, returns a
        TranslateToResponse
        """
        self.routine = routine

    def __call__(self, request: TranslateToRequest) -> TranslateToResponse:
        return self.routine(request)


class TranslateFromFunction:
    def __init__(self, routine):
        """
        :param routine: A translation routine that, given a TranslateToRequest object, returns a
        tuple[TranslateFromResponse, str], the second return value being any potential translation error messages
        """
        self.routine = routine

    def __call__(self, request: TranslateFromRequest) -> tuple[TranslateFromResponse, str]:
        return self.routine(request)


class MonarchTranslator(BaseHTTPRequestHandler):
    translate_to: TranslateToFunction
    translate_from: TranslateFromFunction

    def register_to(self, function: TranslateToFunction):
        """
        :param function: A function that receives a TranslateToRequest object and returns a TranslateToResponse object,
        And any potential errors
        :return:
        """
        self.translate_to = function

    def register_from(self, function: TranslateFromFunction):
        """
        :param function: A function that receives a TranslateFromRequest object and returns a TranslateFromResponse
        object, and any potential errors
        :return:
        """
        self.translate_from = function

    def do_POST(self):
        content_length = int(self.headers.get('Content-Length'))

        post_data = self.rfile.read(content_length)
        data = json.loads(post_data)

        if self.path.startswith("/to"):
            to_request = TranslateToRequest(
                data["agent_id"],
                data["request_id"],
                data["opcode"],
                data["args"]
            )
            to_response = self.translate_to(to_request)
            response_json = {
                "success": to_response.success,
                "error_msg": to_response.error_msg,
                "message": to_response.message
            }
            response = json.dumps(response_json)
            self.send_response(200)
            self.send_header("content-type", "application/json")
            self.end_headers()
            self.wfile.write(bytes(response, "utf-8"))

        elif self.path.startswith("/from"):
            from_request = TranslateFromRequest(
                data["message"]
            )
            from_response, error = self.translate_from(from_request)
            response_json = {
                "success": len(error) == 0,
                "error_msg": error,
                "agent_id": from_response.agent_id,
                "request_id": from_response.request_id,
                "status": from_response.status,
                "responses": from_response.responses,
            }
            response = json.dumps(response_json)
            self.send_response(200)
            self.send_header("content-type", "application/json")
            self.end_headers()
            self.wfile.write(bytes(response, "utf-8"))
        else:
            self.send_response(404)


def translator_service() -> HTTPServer:
    """
    :return: A HTTPServer class using the monarch translator class as a request handler
    """
    service_address = ("localhost", 20000)
    return HTTPServer(service_address, MonarchTranslator)
