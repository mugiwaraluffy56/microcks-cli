import * as vscode from "vscode";
import { MicrocksClient, MicrocksService, MicrocksOperation } from "./microcksClient";

type TreeNode = ServiceItem | OperationItem | MessageItem;

export class ServiceItem extends vscode.TreeItem {
  constructor(public readonly service: MicrocksService) {
    super(
      `${service.name} : ${service.version}`,
      vscode.TreeItemCollapsibleState.Collapsed
    );
    this.tooltip = `${service.type} — ${service.operations?.length ?? 0} operations`;
    this.iconPath = new vscode.ThemeIcon("symbol-interface");
    this.contextValue = "microcksService";
  }
}

export class OperationItem extends vscode.TreeItem {
  constructor(public readonly operation: MicrocksOperation) {
    const label = operation.name.startsWith(operation.method)
      ? operation.name
      : `${operation.method} ${operation.name}`;
    super(
      label,
      vscode.TreeItemCollapsibleState.None
    );
    this.iconPath = new vscode.ThemeIcon("symbol-method");
    this.contextValue = "microcksOperation";
  }
}

export class MessageItem extends vscode.TreeItem {
  constructor(label: string, icon: string) {
    super(label, vscode.TreeItemCollapsibleState.None);
    this.iconPath = new vscode.ThemeIcon(icon);
    this.contextValue = "microcksMessage";
  }
}

export class MicrocksServicesProvider
  implements vscode.TreeDataProvider<TreeNode>
{
  private _onDidChangeTreeData = new vscode.EventEmitter<TreeNode | undefined | void>();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  private services: MicrocksService[] | null = null;
  private error: string | null = null;

  constructor(private client: MicrocksClient) {}

  updateClient(client: MicrocksClient): void {
    this.client = client;
  }

  refresh(): void {
    this.services = null;
    this.error = null;
    this._onDidChangeTreeData.fire();
  }

  getTreeItem(element: TreeNode): vscode.TreeItem {
    return element;
  }

  async getChildren(element?: TreeNode): Promise<TreeNode[]> {
    if (element instanceof ServiceItem) {
      const ops = element.service.operations ?? [];
      if (ops.length === 0) {
        return [new MessageItem("no operations found", "info")];
      }
      return ops.map((op) => new OperationItem(op));
    }

    if (element) {
      return [];
    }

    // root level — fetch services if not cached
    if (this.services === null && this.error === null) {
      try {
        this.services = await this.client.getServices();
      } catch (err: unknown) {
        this.error = err instanceof Error ? err.message : String(err);
      }
    }

    if (this.error) {
      return [new MessageItem(`Could not connect: ${this.error}`, "error")];
    }

    if (!this.services || this.services.length === 0) {
      return [new MessageItem("no services found — import a spec first", "info")];
    }

    return this.services.map((svc) => new ServiceItem(svc));
  }
}
