export interface InaccessibleLink {
  url: string;
  status_code: number;
}

export interface Analysis {
  id: number;
  url: string;
  status: 'queued' | 'running' | 'done' | 'error';
  error_msg?: string;
  page_title?: string;
  html_version?: string;
  heading_counts?: { [key: string]: number };
  internal_link_count: number;
  external_link_count: number;
  inaccessible_links?: InaccessibleLink[];
  has_login_form: boolean;
  created_at: string;
}
