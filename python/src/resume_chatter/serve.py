import logging
import sys
from concurrent import futures
from typing import Any

import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer
from langchain.chains import create_retrieval_chain
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain_community.document_loaders import PyPDFLoader
from langchain_community.vectorstores import FAISS
from langchain_core.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI, OpenAIEmbeddings
from langchain_text_splitters import RecursiveCharacterTextSplitter
from protocgenpy.chat.v1 import chat_pb2, chat_pb2_grpc

logger = logging.getLogger(__name__)


class ChatServicer(chat_pb2_grpc.ChatServiceServicer):
    def __init__(self, *args, **kwargs):
        pass

    def Ask(self, request: Any, context: Any) -> chat_pb2.AskResponse:
        model = "gpt-4o-mini"
        logger.warning(f"using model: {model}")

        llm = ChatOpenAI(model=model)

        loader = PyPDFLoader(request.document_path)
        docs = loader.load()

        text_splitter = RecursiveCharacterTextSplitter()
        documents = text_splitter.split_documents(docs)
        embeddings = OpenAIEmbeddings()
        vector = FAISS.from_documents(documents, embeddings)

        prompt = ChatPromptTemplate.from_template("""Answer the following question based only on the provided context:

        <context>
        {context}
        </context>

        Question: {input}""")

        document_chain = create_stuff_documents_chain(llm, prompt)
        retriever = vector.as_retriever()
        retrieval_chain = create_retrieval_chain(retriever, document_chain)

        return chat_pb2.AskResponse(
            response=str(retrieval_chain.invoke({"input": request.question})['answer'])
        )


def serve() -> None:
    port = "8081"

    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value("SERVING"))

    # GRPC enables SO_REUSEPORT by default.
    #   https://groups.google.com/g/grpc-io/c/RB69llv2tC4
    # This may be recommended for production workloads but causes problems
    # during development when I inevitably launch a second server without
    # killing the first. Disable this behavior for my sanity.
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=10), options=(("grpc.so_reuseport", 0),)
    )
    chat_pb2_grpc.add_ChatServiceServicer_to_server(ChatServicer(), server)  # type: ignore[no-untyped-call]
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port("127.0.0.1:" + port)
    server.start()

    print("1|1|tcp|127.0.0.1:8081|grpc")
    sys.stdout.flush()

    server.wait_for_termination()
