class MockMessage:
    def __init__(self, content=None, tool_calls=None):
        self.content = content
        self.tool_calls = tool_calls


class MockToolCall:
    def __init__(self, name, arguments):
        self.id = "call_mock_id_001"
        self.function = MockFunction(name, arguments)


class MockFunction:
    def __init__(self, name, arguments):
        self.name = name
        self.arguments = arguments


def make_tool_call_response(tool_name, arguments):
    return MockMessage(
        content=None,
        tool_calls=[MockToolCall(tool_name, arguments)],
    )


def make_final_response(content):
    return MockMessage(content=content, tool_calls=None)
